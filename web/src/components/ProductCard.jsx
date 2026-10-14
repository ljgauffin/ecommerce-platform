import React from 'react';
import { ShoppingCart, Package, Scale, AlertCircle, CheckCircle2, Pencil, Trash2 } from 'lucide-react';

export default function ProductCard({ 
  product, 
  cartQuantity, 
  onAddToCart,
  onEditProduct,
  onDeleteProduct,
  showAdminActions = false 
}) {
  const isOutOfStock = product.stock <= 0;
  const isMaxInCart = cartQuantity >= product.stock;

  const renderStockBadge = () => {
    if (isOutOfStock) {
      return (
        <span className="badge badge-rose">
          <AlertCircle size={12} /> Out of stock
        </span>
      );
    }
    if (product.stock <= 5) {
      return (
        <span className="badge badge-amber">
          <AlertCircle size={12} /> Only {product.stock} left
        </span>
      );
    }
    return (
      <span className="badge badge-emerald">
        <CheckCircle2 size={12} /> {product.stock} in stock
      </span>
    );
  };

  return (
    <div 
      className="card-glass"
      style={{
        display: 'flex',
        flexDirection: 'column',
        padding: '20px',
        position: 'relative',
        transition: 'transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease',
      }}
    >
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: 12,
        gap: 8,
      }}>
        <span className="badge badge-cyan" style={{ fontSize: '0.7rem' }}>
          {product.category?.name || (typeof product.category === 'string' ? product.category : 'General')}
        </span>

        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {renderStockBadge()}
          {showAdminActions && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 2, marginLeft: 4 }}>
              <button
                onClick={() => onEditProduct && onEditProduct(product)}
                className="btn btn-ghost"
                style={{ padding: 4, width: 26, height: 26, color: 'var(--text-muted)' }}
                title="Edit product"
              >
                <Pencil size={13} />
              </button>
              <button
                onClick={() => onDeleteProduct && onDeleteProduct(product)}
                className="btn btn-ghost"
                style={{ padding: 4, width: 26, height: 26, color: 'var(--accent-rose)' }}
                title="Delete product"
              >
                <Trash2 size={13} />
              </button>
            </div>
          )}
        </div>
      </div>

      <h3 style={{
        fontSize: '1.05rem',
        marginBottom: 8,
        color: 'var(--text-primary)',
        lineHeight: 1.3,
        minHeight: '2.6em',
      }}>
        {product.name}
      </h3>

      <p style={{
        fontSize: '0.8rem',
        color: 'var(--text-secondary)',
        marginBottom: 16,
        lineHeight: 1.4,
        flexGrow: 1,
        display: '-webkit-box',
        WebkitLineClamp: 2,
        WebkitBoxOrient: 'vertical',
        overflow: 'hidden',
      }}>
        {product.description || 'No description provided.'}
      </p>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        fontSize: '0.75rem',
        color: 'var(--text-muted)',
        marginBottom: 16,
        paddingTop: 12,
        borderTop: '1px solid var(--border-subtle)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Package size={13} />
          <code>{product.sku}</code>
        </div>
        {product.weight_kg > 0 && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <Scale size={13} />
            <span>{product.weight_kg} kg</span>
          </div>
        )}
      </div>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginTop: 'auto',
        gap: 12,
      }}>
        <div>
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>
            Price
          </div>
          <div style={{
            fontSize: '1.3rem',
            fontFamily: 'var(--font-display)',
            fontWeight: 700,
            color: product.price === 0 ? 'var(--accent-emerald)' : 'var(--text-primary)',
          }}>
            {product.price === 0 ? 'FREE' : `$${Number(product.price).toFixed(2)}`}
          </div>
        </div>

        <button
          className={`btn ${isOutOfStock ? 'btn-secondary' : 'btn-primary'}`}
          onClick={() => onAddToCart(product)}
          disabled={isOutOfStock || isMaxInCart}
          style={{ padding: '8px 14px', fontSize: '0.8rem' }}
          id={`add-cart-${product.sku}`}
        >
          <ShoppingCart size={15} />
          <span>
            {isOutOfStock 
              ? 'Out of stock' 
              : isMaxInCart 
                ? 'Max reached' 
                : cartQuantity > 0 
                  ? `In cart (${cartQuantity})` 
                  : 'Add to cart'}
          </span>
        </button>
      </div>
    </div>
  );
}
