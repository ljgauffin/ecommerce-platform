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

  const [products, setProducts] = useState([]);
  const [totalCount, setTotalCount] = useState(0);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Store-specific filters
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("");
  const [inStockOnly, setInStockOnly] = useState(false);
  const [sortBy, setSortBy] = useState("newest");

  // Modals state
  const [isImportOpen, setIsImportOpen] = useState(false);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [isProductFormOpen, setIsProductFormOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState(null);
  const [productToDelete, setProductToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Cart state persisted to localStorage
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

  // Synchronize hash with active tab
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

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();

      // Only apply store search and category filters if we are in the store tab
      if (activeTab === "store") {
        if (searchTerm.trim()) params.set("query", searchTerm.trim());
        if (selectedCategory) params.set("category", selectedCategory);
        if (inStockOnly) params.set("in_stock", "true");
        if (sortBy) params.set("sort", sortBy);
      }

      params.set("limit", "250");

      const response = await fetch(apiUrl(`/api/v1/products?${params.toString()}`));
      if (!response.ok) {
        throw new Error(`Failed to load products: HTTP ${response.status}`);
      }
      const json = await response.json();
      setProducts(json.data || []);
      setTotalCount(
        json.total !== undefined
          ? json.total
          : json.data
            ? json.data.length
            : 0,
      );
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [activeTab, searchTerm, selectedCategory, inStockOnly, sortBy]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchProducts();
    }, 200);
    return () => clearTimeout(timer);
  }, [fetchProducts]);

  // Cart operations
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

      setProducts((prev) => prev.filter((p) => p.id !== productToDelete.id));
      setTotalCount((prev) => Math.max(0, prev - 1));
      setCartItems((prev) =>
        prev.filter((i) => i.product.id !== productToDelete.id),
      );
      addToast(
        `Product "${productToDelete.name}" deleted successfully.`,
        "info",
      );
      fetchCategories();
      setProductToDelete(null);
    } catch (err) {
      addToast(err.message, "error");
    } finally {
      setIsDeleting(false);
    }
  };

  const handleSaveProductSuccess = (savedProduct, isEditing) => {
    fetchProducts();
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
    fetchProducts();
  };

  const handleImportSuccess = (result) => {
    addToast(
      `Import finished: ${result.created} new, ${result.updated} updated, ${result.failed} rejected.`,
      result.failed > 0 ? "info" : "success",
    );
    fetchProducts();
    fetchCategories();
  };

  const totalCartCount = cartItems.reduce((acc, i) => acc + i.quantity, 0);

  return (
    <div
      style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}
    >
      {/* Top Navigation Bar with View Switcher */}
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

      {/* Main Content: Store View OR Product CRUD View */}
      <main style={{ flex: 1 }}>
        {activeTab === "store" ? (
          <StoreView
            products={products}
            totalCount={totalCount}
            loading={loading}
            error={error}
            categories={categories}
            selectedCategory={selectedCategory}
            onSelectCategory={setSelectedCategory}
            inStockOnly={inStockOnly}
            onToggleInStock={setInStockOnly}
            sortBy={sortBy}
            onSelectSort={setSortBy}
            cartItems={cartItems}
            onAddToCart={handleAddToCart}
            onRetry={fetchProducts}
            onClearFilters={() => {
              setSearchTerm("");
              setSelectedCategory("");
              setInStockOnly(false);
            }}
            onNavigateToCrud={() => handleSelectTab("crud")}
          />
        ) : (
          <ProductCrudView
            products={products}
            loading={loading}
            error={error}
            categories={categories}
            onRefresh={() => {
              fetchProducts();
              fetchCategories();
            }}
            onOpenNewProduct={handleOpenNewProduct}
            onOpenEditProduct={handleOpenEditProduct}
            onDeleteProduct={handleDeleteProduct}
            onOpenImport={() => setIsImportOpen(true)}
          />
        )}
      </main>

      {/* Modals & Overlays */}
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
