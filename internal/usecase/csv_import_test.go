package usecase_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) domain.ProductRepository {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Product{}, &domain.Order{}, &domain.OrderItem{}, &domain.IdempotencyRecord{})
	require.NoError(t, err)

	return persistence.NewGormProductRepo(db)
}

func TestCsvImportUseCase_SuccessAndEdgeCases(t *testing.T) {
	repo := setupTestDB(t)
	uc := usecase.NewCsvImportUseCase(repo)
	ctx := context.Background()

	csvContent := `name,sku,description,category,price,stock,weight_kg
Running Shoes,RS-001,Lightweight running shoes for daily training,Footwear,89.99,150,0.35
Organic Coffee Beans,CB-010,"Single origin, medium roast",Food & Beverage,18.75,500,1
Wireless Mouse,WM-042,Ergonomic wireless mouse,Electronics,$29.99,75,0.12
Sticker Pack,SP-001,Free gift,Merchandise,free,1000,0.05
Ultra Monitor 27" 4K,UM-027,High resolution display,Electronics,349.99,30,5.20
<script>alert('xss')</script> Safe Gadget,GAD-001,Testing XSS strip,Gadgets,19.99,20,0.40
Product'; DROP TABLE products; --,SQLI-001,Safe text,Electronics,49.99,10,1.20
`

	result, err := uc.Execute(ctx, strings.NewReader(csvContent))
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 7, result.TotalRows)
	assert.Equal(t, 7, result.Imported)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, result.Errors)

	freeProduct, err := repo.GetBySKU(ctx, "SP-001")
	require.NoError(t, err)
	assert.Equal(t, 0.0, freeProduct.Price)

	mouseProduct, err := repo.GetBySKU(ctx, "WM-042")
	require.NoError(t, err)
	assert.Equal(t, 29.99, mouseProduct.Price)

	xssProduct, err := repo.GetBySKU(ctx, "GAD-001")
	require.NoError(t, err)
	assert.Equal(t, "Safe Gadget", xssProduct.Name)
	assert.NotContains(t, xssProduct.Name, "<script>")

	sqliProduct, err := repo.GetBySKU(ctx, "SQLI-001")
	require.NoError(t, err)
	assert.Contains(t, sqliProduct.Name, "DROP TABLE")
}

func TestCsvImportUseCase_TabDelimited(t *testing.T) {
	repo := setupTestDB(t)
	uc := usecase.NewCsvImportUseCase(repo)
	ctx := context.Background()

	tsvContent := "name\tsku\tdescription\tcategory\tprice\tstock\tweight_kg\n" +
		"Running Shoes\tRS-002\tRunning shoes\tFootwear\t89.99\t150\t0.35\n"

	result, err := uc.Execute(ctx, strings.NewReader(tsvContent))
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, 1, result.Imported)
}

func TestCsvImportUseCase_RowValidationErrors(t *testing.T) {
	repo := setupTestDB(t)
	uc := usecase.NewCsvImportUseCase(repo)
	ctx := context.Background()

	csvContent := `name,sku,description,category,price,stock,weight_kg
Valid Item,VAL-01,Valid,General,10.00,5,1.0
   ,ERR-01,Empty Name,General,10.00,5,1.0
Bad Price,ERR-02,Invalid Price,General,-5.00,5,1.0
Bad Stock,ERR-03,Invalid Stock,General,10.00,-2,1.0
Bad Weight,ERR-04,Invalid Weight,General,10.00,5,
Empty SKU, ,Empty SKU,General,10.00,5,1.0
Another Valid,VAL-02,Valid,General,20.00,10,2.0
`

	result, err := uc.Execute(ctx, strings.NewReader(csvContent))
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 7, result.TotalRows)
	assert.Equal(t, 2, result.Imported)
	assert.Equal(t, 5, result.Failed)
	assert.Len(t, result.Errors, 5)

	val1, err := repo.GetBySKU(ctx, "VAL-01")
	require.NoError(t, err)
	assert.NotNil(t, val1)

	val2, err := repo.GetBySKU(ctx, "VAL-02")
	require.NoError(t, err)
	assert.NotNil(t, val2)
}

func TestCsvImportUseCase_MissingHeader(t *testing.T) {
	repo := setupTestDB(t)
	uc := usecase.NewCsvImportUseCase(repo)
	ctx := context.Background()

	badHeaderCSV := `name,description,category,price
Item,Desc,Cat,10.00
`
	_, err := uc.Execute(ctx, strings.NewReader(badHeaderCSV))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required column")
}

func TestCsvImportUseCase_DuplicateSKUsRejected(t *testing.T) {
	repo := setupTestDB(t)
	uc := usecase.NewCsvImportUseCase(repo)
	ctx := context.Background()

	csvContent := `name,sku,description,category,price,stock,weight_kg
Product A,DUP-01,First occurrence,General,10.00,5,1.0
Product B,UNIQUE-01,Unique item,General,20.00,10,2.0
Product A Clone,DUP-01,Second occurrence,General,15.00,8,1.0
`

	result, err := uc.Execute(ctx, strings.NewReader(csvContent))
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.TotalRows)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 2, result.Failed)
	assert.Len(t, result.Errors, 2)
	assert.Equal(t, 2, result.Errors[0].Row)
	assert.Equal(t, "DUP-01", result.Errors[0].SKU)
	assert.Equal(t, 4, result.Errors[1].Row)
	assert.Equal(t, "DUP-01", result.Errors[1].SKU)

	// DUP-01 must NOT exist in repo
	_, err = repo.GetBySKU(ctx, "DUP-01")
	assert.ErrorIs(t, err, domain.ErrProductNotFound)

	// UNIQUE-01 must exist
	u, err := repo.GetBySKU(ctx, "UNIQUE-01")
	require.NoError(t, err)
	assert.Equal(t, "Unique item", u.Description)
}
