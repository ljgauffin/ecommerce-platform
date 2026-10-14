import React, { useState, useEffect, useCallback } from "react";
import Navbar from "./components/Navbar";
import StoreView from "./components/StoreView";
import ProductCrudView from "./components/ProductCrudView";
import ProductFormModal from "./components/ProductFormModal";
import CsvImportModal from "./components/CsvImportModal";
import DeleteConfirmModal from "./components/DeleteConfirmModal";
import CartDrawer from "./components/CartDrawer";
import Toast from "./components/Toast";
import { apiUrl } from "./config/api";

export default function App() {
  const [activeTab, setActiveTab] = useState(() => {
    if (typeof window !== "undefined") {
      return window.location.hash === "#crud" ? "crud" : "store";
    }
    return "store";
  });

  const [storeProducts, setStoreProducts] = useState([]);
  const [storeTotal, setStoreTotal] = useState(0);
  const [storeLoading, setStoreLoading] = useState(true);
  const [storeLoadingMore, setStoreLoadingMore] = useState(false);
  const [storeHasMore, setStoreHasMore] = useState(true);
  const [storeError, setStoreError] = useState(null);

  const [crudProducts, setCrudProducts] = useState([]);
  const [crudTotal, setCrudTotal] = useState(0);
  const [crudLoading, setCrudLoading] = useState(true);
  const [crudError, setCrudError] = useState(null);
  const [crudPage, setCrudPage] = useState(1);
  const [crudPageSize, setCrudPageSize] = useState(15);
  const [crudSearch, setCrudSearch] = useState("");
  const [crudCategory, setCrudCategory] = useState("");
  const [crudStockFilter, setCrudStockFilter] = useState("all");
  const [crudSortBy, setCrudSortBy] = useState("name-asc");

  const [searchTerm, setSearchTerm] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("");
  const [inStockOnly, setInStockOnly] = useState(false);
  const [sortBy, setSortBy] = useState("newest");

  const [isImportOpen, setIsImportOpen] = useState(false);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [isProductFormOpen, setIsProductFormOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState(null);
  const [productToDelete, setProductToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const [cartItems, setCartItems] = useState(() => {
    try {
      const saved = localStorage.getItem("aether_cart");
      return saved ? JSON.parse(saved) : [];
    } catch {
      return [];
    }
  });

  const [toasts, setToasts] = useState([]);

  const addToast = (message, type = "info") => {
    const id = Date.now() + Math.random().toString(36).substring(2, 6);
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  };

  const removeToast = (id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  useEffect(() => {
    try {
      localStorage.setItem("aether_cart", JSON.stringify(cartItems));
    } catch (e) {
      console.error("Failed to save cart to localStorage", e);
    }
  }, [cartItems]);

  const handleSelectTab = (tab) => {
    setActiveTab(tab);
    if (typeof window !== "undefined") {
      window.location.hash = tab === "crud" ? "#crud" : "#store";
    }
  };

  useEffect(() => {
    const handleHashChange = () => {
      setActiveTab(window.location.hash === "#crud" ? "crud" : "store");
    };
    window.addEventListener("hashchange", handleHashChange);
    return () => window.removeEventListener("hashchange", handleHashChange);
  }, []);

  const [categories, setCategories] = useState([]);

  const fetchCategories = useCallback(async () => {
    try {
      const response = await fetch(apiUrl("/api/v1/products/categories"));
      if (response.ok) {
        const json = await response.json();
        setCategories(json.data || []);
      }
    } catch (err) {
      console.error("Failed to load categories", err);
    }
  }, []);

  const STORE_PAGE_SIZE = 12;

  const fetchInitialStoreProducts = useCallback(async () => {
    setStoreLoading(true);
    setStoreError(null);
    try {
      const params = new URLSearchParams();
      params.set("limit", String(STORE_PAGE_SIZE));
      params.set("offset", "0");
      if (searchTerm.trim()) params.set("query", searchTerm.trim());
      if (selectedCategory) params.set("category", selectedCategory);
      if (inStockOnly) params.set("in_stock", "true");
      if (sortBy) params.set("sort", sortBy);

      const response = await fetch(apiUrl(`/api/v1/products?${params.toString()}`));
      if (!response.ok) {
        throw new Error(`Failed to load products: HTTP ${response.status}`);
      }
      const json = await response.json();
      const list = json.data || [];
      const total = json.total !== undefined ? json.total : list.length;
      setStoreProducts(list);
      setStoreTotal(total);
      setStoreHasMore(list.length < total);
    } catch (err) {
      setStoreError(err.message);
    } finally {
      setStoreLoading(false);
    }
  }, [searchTerm, selectedCategory, inStockOnly, sortBy]);

  const handleLoadMoreStore = useCallback(async () => {
    if (storeLoadingMore || !storeHasMore || storeLoading) return;
    setStoreLoadingMore(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", String(STORE_PAGE_SIZE));
      params.set("offset", String(storeProducts.length));
      if (searchTerm.trim()) params.set("query", searchTerm.trim());
      if (selectedCategory) params.set("category", selectedCategory);
      if (inStockOnly) params.set("in_stock", "true");
      if (sortBy) params.set("sort", sortBy);

      const response = await fetch(apiUrl(`/api/v1/products?${params.toString()}`));
      if (!response.ok) {
        throw new Error(`Failed to load more products: HTTP ${response.status}`);
      }
      const json = await response.json();
      const nextList = json.data || [];
      const total = json.total !== undefined ? json.total : storeProducts.length + nextList.length;

      setStoreProducts((prev) => {
        const existingIds = new Set(prev.map((p) => p.id));
        const uniqueNext = nextList.filter((p) => !existingIds.has(p.id));
        const updated = [...prev, ...uniqueNext];
        setStoreHasMore(updated.length < total && nextList.length > 0);
        return updated;
      });
      setStoreTotal(total);
    } catch (err) {
      console.error("Failed to load more products", err);
    } finally {
      setStoreLoadingMore(false);
    }
  }, [storeLoadingMore, storeHasMore, storeLoading, storeProducts.length, searchTerm, selectedCategory, inStockOnly, sortBy]);

  const fetchCrudProducts = useCallback(async () => {
    setCrudLoading(true);
    setCrudError(null);
    try {
      const params = new URLSearchParams();
      params.set("limit", String(crudPageSize));
      params.set("offset", String((crudPage - 1) * crudPageSize));
      if (crudSearch.trim()) params.set("query", crudSearch.trim());
      if (crudCategory) params.set("category", crudCategory);
      if (crudStockFilter && crudStockFilter !== "all") {
        params.set("stock_status", crudStockFilter);
      }
      if (crudSortBy) params.set("sort", crudSortBy);

      const response = await fetch(apiUrl(`/api/v1/products?${params.toString()}`));
      if (!response.ok) {
        throw new Error(`Failed to load CRUD products: HTTP ${response.status}`);
      }
      const json = await response.json();
      setCrudProducts(json.data || []);
      setCrudTotal(json.total !== undefined ? json.total : (json.data ? json.data.length : 0));
    } catch (err) {
      setCrudError(err.message);
    } finally {
      setCrudLoading(false);
    }
  }, [crudPage, crudPageSize, crudSearch, crudCategory, crudStockFilter, crudSortBy]);

  const handleCrudSearchChange = (term) => {
    setCrudSearch(term);
    setCrudPage(1);
  };

  const handleCrudCategoryChange = (cat) => {
    setCrudCategory(cat);
    setCrudPage(1);
  };

  const handleCrudStockFilterChange = (stock) => {
    setCrudStockFilter(stock);
    setCrudPage(1);
  };

  const handleCrudSortByChange = (sort) => {
    setCrudSortBy(sort);
    setCrudPage(1);
  };

  const handleCrudPageChange = (page) => {
    setCrudPage(page);
  };

  const handleCrudPageSizeChange = (size) => {
    setCrudPageSize(size);
    setCrudPage(1);
  };

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchCrudProducts();
    }, 200);
    return () => clearTimeout(timer);
  }, [fetchCrudProducts]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchInitialStoreProducts();
    }, 200);
    return () => clearTimeout(timer);
  }, [fetchInitialStoreProducts]);

  const handleAddToCart = (product) => {
    setCartItems((prev) => {
      const existing = prev.find((item) => item.product.id === product.id);
      if (existing) {
        if (existing.quantity >= product.stock) {
          addToast(
            `Maximum available stock (${product.stock}) already in cart.`,
            "error",
          );
          return prev;
        }
        addToast(`Added another "${product.name}" to cart.`, "success");
        return prev.map((item) =>
          item.product.id === product.id
            ? { ...item, quantity: item.quantity + 1 }
            : item,
        );
      } else {
        addToast(`"${product.name}" added to cart.`, "success");
        return [...prev, { product, quantity: 1 }];
      }
    });
  };

  const handleUpdateQuantity = (productId, newQuantity) => {
    setCartItems((prev) =>
      prev
        .map((item) => {
          if (item.product.id === productId) {
            const capped = Math.min(newQuantity, item.product.stock);
            return capped > 0 ? { ...item, quantity: capped } : null;
          }
          return item;
        })
        .filter(Boolean),
    );
  };

  const handleRemoveFromCart = (productId) => {
    setCartItems((prev) =>
      prev.filter((item) => item.product.id !== productId),
    );
    addToast("Item removed from cart.", "info");
  };

  const handleClearCart = () => {
    setCartItems([]);
  };


  const handleOpenNewProduct = () => {
    setEditingProduct(null);
    setIsProductFormOpen(true);
  };

  const handleOpenEditProduct = (product) => {
    setEditingProduct(product);
    setIsProductFormOpen(true);
  };

  const handleDeleteProduct = (product) => {
    setProductToDelete(product);
  };

  const handleConfirmDelete = async () => {
    if (!productToDelete) return;
    setIsDeleting(true);

    try {
      const response = await fetch(
        apiUrl(`/api/v1/products/${productToDelete.id}`),
        {
          method: "DELETE",
        },
      );

      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(
          data.error || `HTTP ${response.status} failed to delete product`,
        );
      }

      setCrudProducts((prev) => prev.filter((p) => p.id !== productToDelete.id));
      setCrudTotal((prev) => Math.max(0, prev - 1));
      setStoreProducts((prev) => prev.filter((p) => p.id !== productToDelete.id));
      setStoreTotal((prev) => Math.max(0, prev - 1));
      setCartItems((prev) =>
        prev.filter((i) => i.product.id !== productToDelete.id),
      );
      addToast(
        `Product "${productToDelete.name}" deleted successfully.`,
        "info",
      );
      fetchCategories();
      if (crudProducts.length === 1 && crudPage > 1) {
        setCrudPage((p) => p - 1);
      } else {
        fetchCrudProducts();
      }
      setProductToDelete(null);
    } catch (err) {
      addToast(err.message, "error");
    } finally {
      setIsDeleting(false);
    }
  };

  const handleSaveProductSuccess = (savedProduct, isEditing) => {
    fetchCrudProducts();
    fetchInitialStoreProducts();
    fetchCategories();
    if (isEditing) {
      setCartItems((prev) =>
        prev.map((i) =>
          i.product.id === savedProduct.id
            ? { ...i, product: savedProduct }
            : i,
        ),
      );
      addToast(
        `Product "${savedProduct.name}" updated successfully.`,
        "success",
      );
    } else {
      addToast(
        `Product "${savedProduct.name}" created successfully.`,
        "success",
      );
    }
  };

  const handleCheckoutSuccess = (order) => {
    addToast(`Order ${order.order_number} confirmed!`, "success");
    fetchCrudProducts();
    fetchInitialStoreProducts();
  };

  const handleImportSuccess = (result) => {
    addToast(
      `Import finished: ${result.created} new, ${result.updated} updated, ${result.failed} rejected.`,
      result.failed > 0 ? "info" : "success",
    );
    fetchCrudProducts();
    fetchInitialStoreProducts();
    fetchCategories();
  };

  const totalCartCount = cartItems.reduce((acc, i) => acc + i.quantity, 0);

  return (
    <div
      style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}
    >
      
      <Navbar
        activeTab={activeTab}
        onSelectTab={handleSelectTab}
        searchTerm={searchTerm}
        setSearchTerm={setSearchTerm}
        cartCount={totalCartCount}
        onOpenCart={() => setIsCartOpen(true)}
        onOpenImport={() => setIsImportOpen(true)}
        onOpenNewProduct={handleOpenNewProduct}
      />

      
      <main style={{ flex: 1 }}>
        {activeTab === "store" ? (
          <StoreView
            products={storeProducts}
            totalCount={storeTotal}
            loading={storeLoading}
            error={storeError}
            categories={categories}
            selectedCategory={selectedCategory}
            onSelectCategory={setSelectedCategory}
            inStockOnly={inStockOnly}
            onToggleInStock={setInStockOnly}
            sortBy={sortBy}
            onSelectSort={setSortBy}
            cartItems={cartItems}
            onAddToCart={handleAddToCart}
            onRetry={fetchInitialStoreProducts}
            onClearFilters={() => {
              setSearchTerm("");
              setSelectedCategory("");
              setInStockOnly(false);
            }}
            onNavigateToCrud={() => handleSelectTab("crud")}
            hasMore={storeHasMore}
            loadingMore={storeLoadingMore}
            onLoadMore={handleLoadMoreStore}
          />
        ) : (
          <ProductCrudView
            products={crudProducts}
            totalCount={crudTotal}
            loading={crudLoading}
            error={crudError}
            categories={categories}
            currentPage={crudPage}
            pageSize={crudPageSize}
            onPageChange={handleCrudPageChange}
            onPageSizeChange={handleCrudPageSizeChange}
            searchTerm={crudSearch}
            onSearchChange={handleCrudSearchChange}
            selectedCategory={crudCategory}
            onCategoryChange={handleCrudCategoryChange}
            stockFilter={crudStockFilter}
            onStockFilterChange={handleCrudStockFilterChange}
            sortBy={crudSortBy}
            onSortByChange={handleCrudSortByChange}
            onRefresh={() => {
              fetchCrudProducts();
              fetchCategories();
            }}
            onOpenNewProduct={handleOpenNewProduct}
            onOpenEditProduct={handleOpenEditProduct}
            onDeleteProduct={handleDeleteProduct}
            onOpenImport={() => setIsImportOpen(true)}
          />
        )}
      </main>

      <ProductFormModal
        isOpen={isProductFormOpen}
        onClose={() => setIsProductFormOpen(false)}
        product={editingProduct}
        categories={categories}
        onSaveSuccess={handleSaveProductSuccess}
      />

      <CsvImportModal
        isOpen={isImportOpen}
        onClose={() => setIsImportOpen(false)}
        onImportSuccess={handleImportSuccess}
      />

      <DeleteConfirmModal
        isOpen={Boolean(productToDelete)}
        onClose={() => {
          if (!isDeleting) setProductToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        product={productToDelete}
        loading={isDeleting}
      />

      <CartDrawer
        isOpen={isCartOpen}
        onClose={() => setIsCartOpen(false)}
        cartItems={cartItems}
        onUpdateQuantity={handleUpdateQuantity}
        onRemoveItem={handleRemoveFromCart}
        onClearCart={handleClearCart}
        onCheckoutSuccess={handleCheckoutSuccess}
      />

      <Toast toasts={toasts} onDismiss={removeToast} />
    </div>
  );
}
