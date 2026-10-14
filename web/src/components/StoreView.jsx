import React from "react";
import Filters from "./Filters";
import ProductCard from "./ProductCard";
import {
  PackageOpen,
  RefreshCw,
  AlertCircle,
  ShoppingBag,
  ArrowRight,
} from "lucide-react";

export default function StoreView({
  products,
  totalCount,
  loading,
  error,
  categories,
  selectedCategory,
  onSelectCategory,
  inStockOnly,
  onToggleInStock,
  sortBy,
  onSelectSort,
  cartItems,
  onAddToCart,
  onRetry,
  onClearFilters,
  onNavigateToCrud,
}) {
  return (
    <div>
      {/* Filter and sorting controls */}
      <Filters
        categories={categories}
        selectedCategory={selectedCategory}
        onSelectCategory={onSelectCategory}
        inStockOnly={inStockOnly}
        onToggleInStock={onToggleInStock}
        sortBy={sortBy}
        onSelectSort={onSelectSort}
        totalProducts={totalCount}
      />

      {/* Products Grid / Status */}
      <div
        style={{
          maxWidth: 1300,
          margin: "0 auto",
          padding: "0 24px 60px",
        }}
      >
        {loading ? (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              padding: "100px 0",
              gap: 16,
              color: "var(--text-secondary)",
            }}
          >
            <RefreshCw
              size={36}
              className="animate-spin"
              style={{ color: "var(--accent-cyan)" }}
            />
            <div style={{ fontSize: "0.95rem", fontWeight: 600 }}>
              Loading product catalogue...
            </div>
          </div>
        ) : error ? (
          <div
            style={{
              maxWidth: 500,
              margin: "60px auto",
              padding: "24px",
              backgroundColor: "rgba(244, 63, 94, 0.1)",
              border: "1px solid rgba(244, 63, 94, 0.3)",
              borderRadius: "var(--radius-lg)",
              textAlign: "center",
            }}
          >
            <AlertCircle
              size={32}
              style={{ color: "var(--accent-rose)", margin: "0 auto 12px" }}
            />
            <h3
              style={{
                fontSize: "1.1rem",
                marginBottom: 8,
                color: "var(--accent-rose)",
              }}
            >
              Connection Error
            </h3>
            <p
              style={{
                fontSize: "0.85rem",
                color: "var(--text-secondary)",
                marginBottom: 16,
              }}
            >
              {error}
            </p>
            <button className="btn btn-secondary" onClick={onRetry}>
              <RefreshCw size={14} /> Retry
            </button>
          </div>
        ) : products.length === 0 ? (
          <div
            style={{
              textAlign: "center",
              padding: "80px 20px",
              backgroundColor: "rgba(255, 255, 255, 0.01)",
              borderRadius: "var(--radius-xl)",
              border: "1px dashed var(--border-subtle)",
            }}
          >
            <PackageOpen
              size={48}
              style={{
                color: "var(--text-muted)",
                margin: "0 auto 16px",
                opacity: 0.6,
              }}
            />
            <h3 style={{ fontSize: "1.2rem", marginBottom: 6 }}>
              No products matched your criteria
            </h3>
            <p
              style={{
                fontSize: "0.85rem",
                color: "var(--text-secondary)",
                marginBottom: 20,
              }}
            >
              Try adjusting your search terms or clearing selected category
              filters.
            </p>
            <button className="btn btn-secondary" onClick={onClearFilters}>
              Clear all filters
            </button>
          </div>
        ) : (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
              gap: 20,
            }}
          >
            {products.map((product) => {
              const inCart = cartItems.find((i) => i.product.id === product.id);
              return (
                <ProductCard
                  key={product.id}
                  product={product}
                  cartQuantity={inCart ? inCart.quantity : 0}
                  onAddToCart={onAddToCart}
                  showAdminActions={false}
                />
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
