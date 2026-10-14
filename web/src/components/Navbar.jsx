import React from "react";
import {
  ShoppingBag,
  UploadCloud,
  Search,
  Plus,
  Layers,
  Package,
} from "lucide-react";

export default function Navbar({
  activeTab = "store",
  onSelectTab,
  searchTerm,
  setSearchTerm,
  cartCount,
  onOpenCart,
  onOpenImport,
  onOpenNewProduct,
}) {
  return (
    <header
      style={{
        position: "sticky",
        top: 0,
        zIndex: 100,
        backgroundColor: "var(--bg-glass)",
        backdropFilter: "blur(16px)",
        WebkitBackdropFilter: "blur(16px)",
        borderBottom: "1px solid var(--border-subtle)",
        padding: "12px 24px",
      }}
    >
      <div
        style={{
          maxWidth: 1300,
          margin: "0 auto",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: 16,
          flexWrap: "wrap",
        }}
      >
        
        <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
          <div
            onClick={() => onSelectTab && onSelectTab("store")}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 12,
              cursor: "pointer",
            }}
            title="Go to Store"
          >
            <div
              style={{
                width: 38,
                height: 38,
                borderRadius: "var(--radius-md)",
                background:
                  "linear-gradient(135deg, var(--accent-cyan), var(--accent-blue))",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                boxShadow: "0 4px 12px var(--accent-cyan-glow)",
              }}
            >
              <Layers size={22} color="#fff" />
            </div>
            <div>
              <div
                style={{
                  fontFamily: "var(--font-display)",
                  fontWeight: 800,
                  fontSize: "1.25rem",
                  letterSpacing: "-0.02em",
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                }}
              >
                GilaSW{" "}
                <span style={{ color: "var(--accent-cyan)" }}>Commerce</span>
              </div>
              <div
                style={{
                  fontSize: "0.7rem",
                  color: "var(--text-muted)",
                  textTransform: "uppercase",
                  letterSpacing: "0.05em",
                }}
              >
                Platform
              </div>
            </div>
          </div>

          
          <nav
            style={{
              display: "flex",
              alignItems: "center",
              backgroundColor: "rgba(23, 31, 50, 0.8)",
              padding: 4,
              borderRadius: "var(--radius-lg)",
              border: "1px solid var(--border-subtle)",
              gap: 4,
            }}
          >
            <button
              onClick={() => onSelectTab && onSelectTab("store")}
              id="nav-tab-store"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "7px 16px",
                borderRadius: "var(--radius-md)",
                fontSize: "0.85rem",
                fontWeight: 600,
                cursor: "pointer",
                border: "none",
                transition: "all 0.2s ease",
                backgroundColor:
                  activeTab === "store" ? "var(--accent-cyan)" : "transparent",
                color:
                  activeTab === "store" ? "#0a0d14" : "var(--text-secondary)",
                boxShadow:
                  activeTab === "store"
                    ? "0 2px 10px var(--accent-cyan-glow)"
                    : "none",
              }}
            >
              <ShoppingBag size={16} />
              <span>Store (Shop)</span>
            </button>

            <button
              onClick={() => onSelectTab && onSelectTab("crud")}
              id="nav-tab-crud"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "7px 16px",
                borderRadius: "var(--radius-md)",
                fontSize: "0.85rem",
                fontWeight: 600,
                cursor: "pointer",
                border: "none",
                transition: "all 0.2s ease",
                backgroundColor:
                  activeTab === "crud" ? "var(--accent-cyan)" : "transparent",
                color:
                  activeTab === "crud" ? "#0a0d14" : "var(--text-secondary)",
                boxShadow:
                  activeTab === "crud"
                    ? "0 2px 10px var(--accent-cyan-glow)"
                    : "none",
              }}
            >
              <Package size={16} />
             Product Management (CRUD & CSV)
            </button>
          </nav>
        </div>

        
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            flexWrap: "wrap",
          }}
        >
          {activeTab === "store" ? (
            <>
              
              <div
                style={{
                  position: "relative",
                  minWidth: 260,
                }}
              >
                <Search
                  size={16}
                  style={{
                    position: "absolute",
                    left: 14,
                    top: "50%",
                    transform: "translateY(-50%)",
                    color: "var(--text-muted)",
                  }}
                />
                <input
                  type="text"
                  className="input"
                  placeholder="Search product, SKU, or category..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  style={{
                    paddingLeft: 38,
                    borderRadius: 9999,
                    backgroundColor: "var(--bg-tertiary)",
                    fontSize: "0.82rem",
                    paddingTop: 8,
                    paddingBottom: 8,
                  }}
                />
                {searchTerm && (
                  <button
                    onClick={() => setSearchTerm("")}
                    style={{
                      position: "absolute",
                      right: 12,
                      top: "50%",
                      transform: "translateY(-50%)",
                      background: "none",
                      border: "none",
                      color: "var(--text-muted)",
                      cursor: "pointer",
                      fontSize: 16,
                    }}
                  >
                    ×
                  </button>
                )}
              </div>

              
              <button
                className="btn btn-primary"
                onClick={onOpenCart}
                id="cart-drawer-btn"
                style={{ position: "relative" }}
              >
                <ShoppingBag size={17} />
                <span>Cart</span>
                {cartCount > 0 && (
                  <span
                    style={{
                      backgroundColor: "var(--accent-rose)",
                      color: "#fff",
                      fontSize: "0.7rem",
                      fontWeight: 700,
                      padding: "2px 7px",
                      borderRadius: 9999,
                      marginLeft: 4,
                      boxShadow: "0 2px 6px rgba(244, 63, 94, 0.4)",
                    }}
                  >
                    {cartCount}
                  </span>
                )}
              </button>
            </>
          ) : (
            <>
              
              <button
                className="btn btn-secondary"
                onClick={onOpenNewProduct}
                id="nav-new-product-btn"
                style={{ fontSize: "0.85rem" }}
              >
                <Plus size={16} style={{ color: "var(--accent-cyan)" }} />
                <span>New Product</span>
              </button>

              <button
                className="btn btn-secondary"
                onClick={onOpenImport}
                id="nav-import-csv-btn"
                style={{
                  fontSize: "0.85rem",
                  borderColor: "rgba(6, 182, 212, 0.4)",
                  color: "var(--accent-cyan)",
                }}
              >
                <UploadCloud size={16} />
                <span>Import CSV</span>
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
