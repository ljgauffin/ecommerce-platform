import React, { useState, useEffect } from 'react';
import { Package, X, Check, AlertTriangle, RefreshCw } from 'lucide-react';
import { apiUrl } from '../config/api';

export default function ProductFormModal({ isOpen, onClose, product, onSaveSuccess, categories = [] }) {
  const isEditing = Boolean(product && product.id);

  const [formData, setFormData] = useState({
    sku: '',
    name: '',
    description: '',
    category: '',
    price: '',
    stock: '',
    weight_kg: '',
  });

  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (product) {
      setFormData({
        sku: product.sku || '',
        name: product.name || '',
        description: product.description || '',
        category: product.category || '',
        price: product.price !== undefined ? product.price : '',
        stock: product.stock !== undefined ? product.stock : '',
        weight_kg: product.weight_kg !== undefined ? product.weight_kg : '',
      });
    } else {
      setFormData({
        sku: '',
        name: '',
        description: '',
        category: '',
        price: '',
        stock: '',
        weight_kg: '',
      });
    }
    setErrorMsg('');
  }, [product, isOpen]);

  if (!isOpen) return null;

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrorMsg('');

    if (!formData.sku.trim()) {
      setErrorMsg('SKU is required');
      return;
    }
    if (!formData.name.trim()) {
      setErrorMsg('Product name is required');
      return;
    }
    if (!formData.category.trim()) {
      setErrorMsg('Category is required');
      return;
    }

    const priceNum = parseFloat(formData.price);
    if (isNaN(priceNum) || priceNum < 0) {
      setErrorMsg('Price must be a valid non-negative number');
      return;
    }

    const stockNum = parseInt(formData.stock, 10);
    if (isNaN(stockNum) || stockNum < 0) {
      setErrorMsg('Stock must be a non-negative integer');
      return;
    }

    const weightNum = parseFloat(formData.weight_kg);
    if (isNaN(weightNum) || weightNum < 0) {
      setErrorMsg('Weight must be a non-negative number');
      return;
    }

    setLoading(true);

    const payload = {
      sku: formData.sku.trim(),
      name: formData.name.trim(),
      description: formData.description.trim(),
      category: formData.category.trim(),
      price: priceNum,
      stock: stockNum,
      weight_kg: weightNum,
    };

    try {
      const url = isEditing
        ? apiUrl(`/api/v1/products/${product.id}`)
        : apiUrl('/api/v1/products');
      const method = isEditing ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to save product');
      }

      onSaveSuccess(data, isEditing);
      onClose();
    } catch (err) {
      setErrorMsg(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div 
        className="modal-container" 
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: 540 }}
      >
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Package size={22} style={{ color: 'var(--accent-cyan)' }} />
            <h2 style={{ fontSize: '1.2rem' }}>
              {isEditing ? 'Edit Product' : 'Create New Product'}
            </h2>
          </div>
          <button className="btn btn-ghost" onClick={onClose} style={{ padding: 6, borderRadius: '50%' }}>
            <X size={18} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="modal-body" style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            {errorMsg && (
              <div style={{
                backgroundColor: 'rgba(244, 63, 94, 0.12)',
                border: '1px solid rgba(244, 63, 94, 0.3)',
                color: 'var(--accent-rose)',
                padding: '10px 14px',
                borderRadius: 'var(--radius-md)',
                fontSize: '0.85rem',
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}>
                <AlertTriangle size={16} />
                <span>{errorMsg}</span>
              </div>
            )}

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                  SKU *
                </label>
                <input
                  type="text"
                  name="sku"
                  required
                  disabled={isEditing}
                  className="input"
                  placeholder="e.g. RS-001"
                  value={formData.sku}
                  onChange={handleChange}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                  Category *
                </label>
                <select
                  name="category"
                  required
                  className="select"
                  value={formData.category}
                  onChange={handleChange}
                  style={{ width: '100%', height: 38 }}
                >
                  <option value="">Select category...</option>
                  {categories.map((c) => {
                    const val = typeof c === 'string' ? c : c.name;
                    return (
                      <option key={val} value={val}>
                        {val}
                      </option>
                    );
                  })}
                </select>
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                Product Name *
              </label>
              <input
                type="text"
                name="name"
                required
                className="input"
                placeholder="Product title"
                value={formData.name}
                onChange={handleChange}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                Description
              </label>
              <textarea
                name="description"
                rows="3"
                className="input"
                placeholder="Optional description..."
                style={{ resize: 'vertical' }}
                value={formData.description}
                onChange={handleChange}
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                  Price ($) *
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  name="price"
                  required
                  className="input"
                  placeholder="0.00"
                  value={formData.price}
                  onChange={handleChange}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                  Stock *
                </label>
                <input
                  type="number"
                  min="0"
                  name="stock"
                  required
                  className="input"
                  placeholder="0"
                  value={formData.stock}
                  onChange={handleChange}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>
                  Weight (kg) *
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  name="weight_kg"
                  required
                  className="input"
                  placeholder="0.00"
                  value={formData.weight_kg}
                  onChange={handleChange}
                />
              </div>
            </div>
          </div>

          <div className="modal-footer">
            <button type="button" className="btn btn-ghost" onClick={onClose} disabled={loading}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? (
                <>
                  <RefreshCw size={15} className="animate-spin" />
                  <span>Saving...</span>
                </>
              ) : (
                <>
                  <Check size={15} />
                  <span>{isEditing ? 'Save Changes' : 'Create Product'}</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
