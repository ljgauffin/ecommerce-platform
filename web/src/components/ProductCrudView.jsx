import React, { useState, useMemo } from "react";
import {
  Package,
  Plus,
  UploadCloud,
  RefreshCw,
  Search,
  Pencil,
  Trash2,
  Copy,
  Check,
  AlertTriangle,
  AlertCircle,
  CheckCircle2,
  Boxes,
  Scale,
  FileSpreadsheet,
  Download,
  Filter,
  ArrowUpDown,
} from "lucide-react";

export default function ProductCrudView({
  products,
  loading,
  error,
  categories,
  onRefresh,
  onOpenNewProduct,
  onOpenEditProduct,
  onDeleteProduct,
  onOpenImport,
}) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("");
  const [stockFilter, setStockFilter] = useState("all"); // all, in_stock, low_stock, out_of_stock
  const [sortBy, setSortBy] = useState("name-asc");
  const [copiedSku, setCopiedSku] = useState(null);
  const [showCsvHelp, setShowCsvHelp] = useState(false);

  // Copy SKU helper with toast-like feedback
  const handleCopySku = (sku, e) => {
    e.stopPropagation();
    navigator.clipboard.writeText(sku).then(() => {
      setCopiedSku(sku);
      setTimeout(() => setCopiedSku(null), 2000);
    });
  };

  // KPIs
  const totalProducts = products.length;
  const totalStock = useMemo(() => {
    return products.reduce((acc, p) => acc + (Number(p.stock) || 0), 0);
  }, [products]);

  const outOfStockCount = useMemo(() => {
    return products.filter((p) => p.stock <= 0).length;
  }, [products]);

  const lowStockCount = useMemo(() => {
    return products.filter((p) => p.stock > 0 && p.stock <= 5).length;
  }, [products]);

  // Filter & sort logic
  const filteredProducts = useMemo(() => {
    return products
      .filter((p) => {
        // Search filter
        if (searchTerm.trim()) {
          const q = searchTerm.toLowerCase().trim();
          const matchSku = p.sku?.toLowerCase().includes(q);
          const matchName = p.name?.toLowerCase().includes(q);
          const matchCategory = p.category?.toLowerCase().includes(q);
          const matchDesc = p.description?.toLowerCase().includes(q);
          if (!matchSku && !matchName && !matchCategory && !matchDesc) {
            return false;
          }
        }

        // Category filter
        if (selectedCategory && p.category !== selectedCategory) {
          return false;
        }

        // Stock filter
        if (stockFilter === "in_stock" && p.stock <= 0) return false;
        if (stockFilter === "low_stock" && (p.stock <= 0 || p.stock > 5))
          return false;
        if (stockFilter === "out_of_stock" && p.stock > 0) return false;

        return true;
      })
      .sort((a, b) => {
        switch (sortBy) {
          case "name-asc":
            return a.name.localeCompare(b.name);
          case "name-desc":
            return b.name.localeCompare(a.name);
          case "sku-asc":
            return a.sku.localeCompare(b.sku);
          case "price-asc":
            return a.price - b.price;
          case "price-desc":
            return b.price - a.price;
          case "stock-asc":
            return a.stock - b.stock;
          case "stock-desc":
            return b.stock - a.stock;
          default:
            return 0;
        }
      });
  }, [products, searchTerm, selectedCategory, stockFilter, sortBy]);

  // Template download
  const handleDownloadCsvTemplate = () => {
    const csvContent =
      "sku,name,description,category,price,stock,weight_kg\n" +
      "RS-001,Sneaker Runner Pro,Lightweight road running shoes,Footwear,129.99,50,0.85\n" +
      "RS-002,Active Dry T-Shirt,Breathable polyester workout tee,Apparel,34.50,120,0.22\n" +
      "RS-003,Hydration Bottle 750ml,Insulated stainless steel sports bottle,Accessories,24.00,85,0.40";

    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("download", "products_template.csv");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div
      style={{ maxWidth: 1300, margin: "0 auto", padding: "24px 24px 60px" }}
    >
      <div
        className="card-glass"
        style={{
          padding: "16px 20px",
          marginBottom: 20,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          flexWrap: "wrap",
          gap: 16,
        }}
      >
        {/* Search Input */}
        <div style={{ flex: "1 1 260px", position: "relative" }}>
          <Search
            size={16}
            style={{
              position: "absolute",
              left: 12,
              top: "50%",
              transform: "translateY(-50%)",
              color: "var(--text-muted)",
            }}
          />
          <input
            type="text"
            className="input"
            placeholder="Filter by SKU, name, category, or description..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ paddingLeft: 38, fontSize: "0.85rem" }}
          />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm("")}
              style={{
                position: "absolute",
                right: 10,
                top: "50%",
                transform: "translateY(-50%)",
                background: "none",
                border: "none",
                color: "var(--text-muted)",
                cursor: "pointer",
              }}
            >
              ×
            </button>
          )}
        </div>

        {/* Filters and Sorters */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            flexWrap: "wrap",
          }}
        >
          {/* Category Filter */}
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <span style={{ fontSize: "0.8rem", color: "var(--text-muted)" }}>
              Category:
            </span>
            <select
              className="select"
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              style={{ width: "auto", padding: "8px 12px", fontSize: "0.8rem" }}
            >
              <option value="">All ({categories.length})</option>
              {categories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </select>
          </div>

          {/* Stock Filter */}
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <span style={{ fontSize: "0.8rem", color: "var(--text-muted)" }}>
              Stock:
            </span>
            <select
              className="select"
              value={stockFilter}
              onChange={(e) => setStockFilter(e.target.value)}
              style={{ width: "auto", padding: "8px 12px", fontSize: "0.8rem" }}
            >
              <option value="all">All statuses</option>
              <option value="in_stock">In stock (&gt; 0)</option>
              <option value="low_stock">Low stock (1-5)</option>
              <option value="out_of_stock">Out of stock (0)</option>
            </select>
          </div>

          {/* Sorter */}
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <ArrowUpDown size={15} style={{ color: "var(--text-muted)" }} />
            <select
              className="select"
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              style={{ width: "auto", padding: "8px 12px", fontSize: "0.8rem" }}
            >
              <option value="name-asc">Name (A-Z)</option>
              <option value="name-desc">Name (Z-A)</option>
              <option value="sku-asc">SKU (A-Z)</option>
              <option value="price-asc">Price: Low to High</option>
              <option value="price-desc">Price: High to Low</option>
              <option value="stock-desc">Highest Stock</option>
              <option value="stock-asc">Lowest Stock</option>
            </select>
          </div>
        </div>
      </div>

      {/* ABM Products Table */}
      <div
        className="card-glass"
        style={{ overflow: "hidden", border: "1px solid var(--border-subtle)" }}
      >
        {loading ? (
          <div
            style={{
              padding: "80px 20px",
              textAlign: "center",
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              gap: 14,
              color: "var(--text-secondary)",
            }}
          >
            <RefreshCw
              size={32}
              className="animate-spin"
              style={{ color: "var(--accent-cyan)" }}
            />
            <span style={{ fontSize: "0.9rem" }}>
              Loading catalogue from database...
            </span>
          </div>
        ) : error ? (
          <div style={{ padding: "60px 20px", textAlign: "center" }}>
            <AlertCircle
              size={32}
              style={{ color: "var(--accent-rose)", margin: "0 auto 10px" }}
            />
            <div
              style={{
                color: "var(--accent-rose)",
                fontWeight: 600,
                marginBottom: 8,
              }}
            >
              {error}
            </div>
            <button className="btn btn-secondary" onClick={onRefresh}>
              <RefreshCw size={14} /> Retry
            </button>
          </div>
        ) : filteredProducts.length === 0 ? (
          <div
            style={{
              padding: "70px 20px",
              textAlign: "center",
              color: "var(--text-secondary)",
            }}
          >
            <Package
              size={44}
              style={{
                color: "var(--text-muted)",
                margin: "0 auto 14px",
                opacity: 0.6,
              }}
            />
            <h3
              style={{
                fontSize: "1.15rem",
                color: "var(--text-primary)",
                marginBottom: 6,
              }}
            >
              No products found
            </h3>
            <p style={{ fontSize: "0.85rem", marginBottom: 20 }}>
              {products.length === 0
                ? "The catalogue is currently empty. Get started by creating a product or performing a CSV import."
                : "No products matched the current search criteria."}
            </p>
            <div style={{ display: "flex", gap: 12, justifyContent: "center" }}>
              {products.length > 0 && (
                <button
                  className="btn btn-secondary"
                  onClick={() => {
                    setSearchTerm("");
                    setSelectedCategory("");
                    setStockFilter("all");
                  }}
                >
                  Reset filters
                </button>
              )}
              <button className="btn btn-primary" onClick={onOpenNewProduct}>
                <Plus size={16} /> Create Product
              </button>
              <button className="btn btn-secondary" onClick={onOpenImport}>
                <UploadCloud size={16} /> Import CSV
              </button>
            </div>
          </div>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table
              style={{
                width: "100%",
                borderCollapse: "collapse",
                fontSize: "0.85rem",
                textAlign: "left",
              }}
            >
              <thead>
                <tr
                  style={{
                    backgroundColor: "rgba(10, 13, 20, 0.7)",
                    borderBottom: "1px solid var(--border-subtle)",
                    color: "var(--text-muted)",
                    fontSize: "0.75rem",
                    textTransform: "uppercase",
                    letterSpacing: "0.04em",
                  }}
                >
                  <th style={{ padding: "14px 18px", width: 140 }}>SKU</th>
                  <th style={{ padding: "14px 18px" }}>Product</th>
                  <th style={{ padding: "14px 18px", width: 150 }}>Category</th>
                  <th style={{ padding: "14px 18px", width: 110 }}>Price</th>
                  <th style={{ padding: "14px 18px", width: 140 }}>Stock</th>
                  <th style={{ padding: "14px 18px", width: 100 }}>Weight</th>
                  <th
                    style={{
                      padding: "14px 18px",
                      width: 160,
                      textAlign: "right",
                    }}
                  >
                    Actions (CRUD)
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredProducts.map((p) => {
                  const isOutOfStock = p.stock <= 0;
                  const isLowStock = p.stock > 0 && p.stock <= 5;

                  return (
                    <tr
                      key={p.id}
                      style={{
                        borderBottom: "1px solid var(--border-subtle)",
                        transition: "background-color 0.15s ease",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.backgroundColor =
                          "rgba(255, 255, 255, 0.02)")
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.backgroundColor = "transparent")
                      }
                    >
                      {/* SKU */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                        }}
                      >
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: 6,
                          }}
                        >
                          <code
                            style={{
                              fontFamily: "monospace",
                              fontSize: "0.8rem",
                              padding: "3px 7px",
                              backgroundColor: "rgba(6, 182, 212, 0.1)",
                              border: "1px solid rgba(6, 182, 212, 0.25)",
                              borderRadius: "var(--radius-sm)",
                              color: "var(--accent-cyan)",
                            }}
                          >
                            {p.sku}
                          </code>
                          <button
                            onClick={(e) => handleCopySku(p.sku, e)}
                            className="btn btn-ghost"
                            style={{
                              padding: 4,
                              width: 22,
                              height: 22,
                              color: "var(--text-muted)",
                            }}
                            title="Copy SKU"
                          >
                            {copiedSku === p.sku ? (
                              <Check
                                size={12}
                                style={{ color: "var(--accent-emerald)" }}
                              />
                            ) : (
                              <Copy size={12} />
                            )}
                          </button>
                        </div>
                      </td>

                      {/* Product Name & Description */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                        }}
                      >
                        <div
                          style={{
                            fontWeight: 600,
                            color: "var(--text-primary)",
                            marginBottom: 2,
                          }}
                        >
                          {p.name}
                        </div>
                        {p.description && (
                          <div
                            style={{
                              fontSize: "0.75rem",
                              color: "var(--text-secondary)",
                              maxWidth: 380,
                              whiteSpace: "nowrap",
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                            }}
                          >
                            {p.description}
                          </div>
                        )}
                      </td>

                      {/* Category */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                        }}
                      >
                        <span
                          className="badge badge-cyan"
                          style={{ fontSize: "0.72rem" }}
                        >
                          {p.category || "General"}
                        </span>
                      </td>

                      {/* Price */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                          fontWeight: 600,
                        }}
                      >
                        <span
                          style={{
                            color:
                              p.price === 0
                                ? "var(--accent-emerald)"
                                : "var(--text-primary)",
                          }}
                        >
                          {p.price === 0
                            ? "FREE"
                            : `$${Number(p.price).toFixed(2)}`}
                        </span>
                      </td>

                      {/* Stock Badge */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                        }}
                      >
                        {isOutOfStock ? (
                          <span
                            className="badge badge-rose"
                            style={{ fontSize: "0.72rem" }}
                          >
                            <AlertCircle size={12} /> Out of stock (0)
                          </span>
                        ) : isLowStock ? (
                          <span
                            className="badge badge-amber"
                            style={{ fontSize: "0.72rem" }}
                          >
                            <AlertTriangle size={12} /> Low stock ({p.stock})
                          </span>
                        ) : (
                          <span
                            className="badge badge-emerald"
                            style={{ fontSize: "0.72rem" }}
                          >
                            <CheckCircle2 size={12} /> {p.stock} in stock
                          </span>
                        )}
                      </td>

                      {/* Weight */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                          color: "var(--text-secondary)",
                        }}
                      >
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: 4,
                            fontSize: "0.8rem",
                          }}
                        >
                          <Scale
                            size={13}
                            style={{ color: "var(--text-muted)" }}
                          />
                          <span>
                            {p.weight_kg ? `${p.weight_kg} kg` : "0 kg"}
                          </span>
                        </div>
                      </td>

                      {/* CRUD Actions */}
                      <td
                        style={{
                          padding: "14px 18px",
                          verticalAlign: "middle",
                          textAlign: "right",
                        }}
                      >
                        <div
                          style={{
                            display: "inline-flex",
                            alignItems: "center",
                            gap: 6,
                          }}
                        >
                          <button
                            onClick={() => onOpenEditProduct(p)}
                            className="btn btn-secondary"
                            style={{ padding: "6px 10px", fontSize: "0.75rem" }}
                            title="Edit product"
                            id={`edit-product-${p.sku}`}
                          >
                            <Pencil
                              size={13}
                              style={{ color: "var(--accent-cyan)" }}
                            />
                            <span>Edit</span>
                          </button>
                          <button
                            onClick={() => onDeleteProduct(p)}
                            className="btn btn-danger"
                            style={{ padding: "6px 10px", fontSize: "0.75rem" }}
                            title="Delete product"
                            id={`delete-product-${p.sku}`}
                          >
                            <Trash2 size={13} />
                            <span>Delete</span>
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Table Footer info */}
        {!loading && !error && filteredProducts.length > 0 && (
          <div
            style={{
              padding: "12px 18px",
              backgroundColor: "rgba(10, 13, 20, 0.4)",
              borderTop: "1px solid var(--border-subtle)",
              fontSize: "0.75rem",
              color: "var(--text-muted)",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              flexWrap: "wrap",
              gap: 12,
            }}
          >
            <div>
              Showing <strong>{filteredProducts.length}</strong> of{" "}
              <strong>{totalProducts}</strong> registered products
              {searchTerm && (
                <span>
                  {" "}
                  matching "<em>{searchTerm}</em>"
                </span>
              )}
            </div>
            <div style={{ display: "flex", gap: 16 }}>
              <span>
                Stock in view:{" "}
                <strong>
                  {filteredProducts.reduce((acc, p) => acc + (p.stock || 0), 0)}
                </strong>{" "}
                units
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
