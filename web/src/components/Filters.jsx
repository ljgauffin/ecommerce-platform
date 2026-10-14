import React from 'react';
import { ArrowDownUp } from 'lucide-react';

export default function Filters({
  categories,
  selectedCategory,
  onSelectCategory,
  inStockOnly,
  onToggleInStock,
  sortBy,
  onSelectSort,
  totalProducts
}) {
  return (
    <div style={{
      maxWidth: 1300,
      margin: '24px auto 16px',
      padding: '0 24px',
      display: 'flex',
      flexDirection: 'column',
      gap: 16,
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: 16,
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          overflowX: 'auto',
          paddingBottom: 4,
          scrollbarWidth: 'none',
        }}>
          <button
            onClick={() => onSelectCategory('')}
            className={`btn ${selectedCategory === '' ? 'btn-primary' : 'btn-secondary'}`}
            style={{ padding: '6px 14px', fontSize: '0.8rem', borderRadius: 9999 }}
          >
            All Products
          </button>
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => onSelectCategory(cat)}
              className={`btn ${selectedCategory === cat ? 'btn-primary' : 'btn-secondary'}`}
              style={{ padding: '6px 14px', fontSize: '0.8rem', borderRadius: 9999 }}
            >
              {cat}
            </button>
          ))}
        </div>

        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 16,
          marginLeft: 'auto',
        }}>
          <label style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            cursor: 'pointer',
            fontSize: '0.85rem',
            color: 'var(--text-secondary)',
            userSelect: 'none',
          }}>
            <input
              type="checkbox"
              checked={inStockOnly}
              onChange={(e) => onToggleInStock(e.target.checked)}
              style={{
                accentColor: 'var(--accent-cyan)',
                width: 16,
                height: 16,
                cursor: 'pointer',
              }}
            />
            <span>In stock only</span>
          </label>

          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <ArrowDownUp size={15} style={{ color: 'var(--text-muted)' }} />
            <select
              value={sortBy}
              onChange={(e) => onSelectSort(e.target.value)}
              className="select"
              style={{
                padding: '6px 12px',
                fontSize: '0.8rem',
                width: 'auto',
                borderRadius: 'var(--radius-sm)',
              }}
            >
              <option value="newest">Newest first</option>
              <option value="price-asc">Price: Low to High</option>
              <option value="price-desc">Price: High to Low</option>
              <option value="name">Name (A-Z)</option>
            </select>
          </div>
        </div>
      </div>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        fontSize: '0.8rem',
        color: 'var(--text-muted)',
      }}>
        <span>
          Showing <strong>{totalProducts}</strong> products
          {selectedCategory && <span> in <em>{selectedCategory}</em></span>}
          {inStockOnly && <span> (available stock only)</span>}
        </span>
      </div>
    </div>
  );
}
