import React, { useState } from 'react';
import { ShoppingBag, Trash2, Plus, Minus, CheckCircle2, AlertCircle, X, ArrowRight, ShieldCheck, RefreshCw } from 'lucide-react';
import { apiUrl } from '../config/api';

export default function CartDrawer({
  isOpen,
  onClose,
  cartItems,
  onUpdateQuantity,
  onRemoveItem,
  onClearCart,
  onCheckoutSuccess
}) {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [confirmedOrder, setConfirmedOrder] = useState(null);

  if (!isOpen) return null;

  const totalAmount = cartItems.reduce(
    (sum, item) => sum + item.product.price * item.quantity,
    0
  );

  const handleCheckout = async (e) => {
    e.preventDefault();
    setErrorMessage('');

    if (!email || !email.includes('@')) {
      setErrorMessage('Please enter a valid customer email address.');
      return;
    }

    if (cartItems.length === 0) {
      setErrorMessage('Your cart is empty.');
      return;
    }

    setLoading(true);

    const idempotencyKey = typeof crypto !== 'undefined' && crypto.randomUUID 
      ? crypto.randomUUID() 
      : 'idemp-' + Math.random().toString(36).substring(2, 15) + Date.now();

    const payload = {
      customer_email: email.trim(),
      idempotency_key: idempotencyKey,
      items: cartItems.map((item) => ({
        product_id: item.product.id,
        quantity: item.quantity,
      })),
    };

    try {
      const response = await fetch(apiUrl('/api/v1/orders/checkout'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Idempotency-Key': idempotencyKey,
        },
        body: JSON.stringify(payload),
      });

      const data = await response.json();

      if (!response.ok) {
        if (response.status === 409) {
          throw new Error('One or more items in your cart exceed available stock. Please reduce quantity or remove out-of-stock items.');
        }
        throw new Error(data.error || 'Failed to complete checkout');
      }

      setConfirmedOrder(data);
      onClearCart();
      if (onCheckoutSuccess) {
        onCheckoutSuccess(data);
      }
    } catch (err) {
      setErrorMessage(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setConfirmedOrder(null);
    setErrorMessage('');
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={handleClose}>
      <div 
        className="modal-container"
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: 560 }}
      >
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <ShoppingBag size={22} style={{ color: 'var(--accent-cyan)' }} />
            <h2 style={{ fontSize: '1.2rem' }}>
              {confirmedOrder ? 'Order Confirmation' : 'Shopping Cart'}
            </h2>
          </div>
          <button 
            className="btn btn-ghost" 
            onClick={handleClose}
            style={{ padding: 6, borderRadius: '50%' }}
          >
            <X size={18} />
          </button>
        </div>

        <div className="modal-body">
          {confirmedOrder ? (
            <div style={{ textAlign: 'center', padding: '16px 8px' }}>
              <div style={{
                width: 64,
                height: 64,
                borderRadius: '50%',
                backgroundColor: 'rgba(16, 185, 129, 0.12)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 16px',
                color: 'var(--accent-emerald)',
                boxShadow: '0 0 24px var(--accent-emerald-glow)',
              }}>
                <CheckCircle2 size={36} />
              </div>

              <h3 style={{ fontSize: '1.4rem', marginBottom: 6 }}>
                Order Placed Successfully!
              </h3>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: 24 }}>
                Stock was atomically deducted and order has been registered.
              </p>

              <div style={{
                backgroundColor: 'var(--bg-tertiary)',
                borderRadius: 'var(--radius-lg)',
                padding: '20px',
                textAlign: 'left',
                border: '1px solid var(--border-subtle)',
                marginBottom: 24,
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Order Number</span>
                  <code style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--accent-cyan)' }}>
                    {confirmedOrder.order_number}
                  </code>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Customer</span>
                  <span style={{ fontSize: '0.85rem', fontWeight: 600 }}>
                    {confirmedOrder.customer_email}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Payment Ref</span>
                  <code style={{ fontSize: '0.8rem' }}>
                    {confirmedOrder.payment_reference}
                  </code>
                </div>
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  paddingTop: 12,
                  borderTop: '1px solid var(--border-subtle)',
                  fontSize: '1rem',
                  fontWeight: 700,
                }}>
                  <span>Total Amount</span>
                  <span style={{ color: 'var(--accent-emerald)' }}>
                    ${Number(confirmedOrder.total_amount).toFixed(2)}
                  </span>
                </div>
              </div>

              <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                fontSize: '0.75rem',
                color: 'var(--text-muted)',
              }}>
                <ShieldCheck size={14} style={{ color: 'var(--accent-emerald)' }} />
                <span>Idempotent transaction verified</span>
              </div>
            </div>
          ) : (
            <div>
              {errorMessage && (
                <div style={{
                  backgroundColor: 'rgba(244, 63, 94, 0.12)',
                  border: '1px solid rgba(244, 63, 94, 0.3)',
                  color: 'var(--accent-rose)',
                  padding: '12px 16px',
                  borderRadius: 'var(--radius-md)',
                  marginBottom: 16,
                  fontSize: '0.85rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                }}>
                  <AlertCircle size={16} />
                  <span>{errorMessage}</span>
                </div>
              )}

              {cartItems.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '40px 0' }}>
                  <ShoppingBag size={48} style={{ color: 'var(--text-muted)', marginBottom: 12, opacity: 0.5 }} />
                  <div style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>
                    Your cart is currently empty
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: 4 }}>
                    Add some products from the catalog to proceed
                  </div>
                </div>
              ) : (
                <>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 24 }}>
                    {cartItems.map(({ product, quantity }) => (
                      <div 
                        key={product.id}
                        style={{
                          backgroundColor: 'var(--bg-tertiary)',
                          borderRadius: 'var(--radius-md)',
                          padding: '12px 16px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          gap: 12,
                        }}
                      >
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{
                            fontWeight: 600,
                            fontSize: '0.9rem',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                          }}>
                            {product.name}
                          </div>
                          <div style={{
                            fontSize: '0.75rem',
                            color: 'var(--text-muted)',
                            display: 'flex',
                            gap: 8,
                          }}>
                            <code>{product.sku}</code>
                            <span>•</span>
                            <span>${Number(product.price).toFixed(2)} each</span>
                            <span>•</span>
                            <span>Max stock: {product.stock}</span>
                          </div>
                        </div>

                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <button
                            className="btn btn-ghost"
                            onClick={() => onUpdateQuantity(product.id, quantity - 1)}
                            disabled={quantity <= 1}
                            style={{ padding: 4, width: 28, height: 28 }}
                          >
                            <Minus size={13} />
                          </button>
                          <span style={{ fontSize: '0.85rem', fontWeight: 600, minWidth: 20, textAlign: 'center' }}>
                            {quantity}
                          </span>
                          <button
                            className="btn btn-ghost"
                            onClick={() => onUpdateQuantity(product.id, quantity + 1)}
                            disabled={quantity >= product.stock}
                            style={{ padding: 4, width: 28, height: 28 }}
                          >
                            <Plus size={13} />
                          </button>
                        </div>

                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                          <div style={{ fontWeight: 700, fontSize: '0.95rem', minWidth: 60, textAlign: 'right' }}>
                            ${(product.price * quantity).toFixed(2)}
                          </div>
                          <button
                            className="btn btn-ghost"
                            onClick={() => onRemoveItem(product.id)}
                            style={{ padding: 6, color: 'var(--accent-rose)' }}
                            title="Remove item"
                          >
                            <Trash2 size={15} />
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>

                  <form onSubmit={handleCheckout} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                    <div>
                      <label style={{
                        display: 'block',
                        fontSize: '0.8rem',
                        fontWeight: 600,
                        color: 'var(--text-secondary)',
                        marginBottom: 6,
                      }}>
                        Customer Email (Required for receipt & order tracking)
                      </label>
                      <input
                        type="email"
                        required
                        className="input"
                        placeholder="buyer@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        id="checkout-email-input"
                      />
                    </div>

                    <div style={{
                      backgroundColor: 'rgba(255, 255, 255, 0.02)',
                      border: '1px solid var(--border-subtle)',
                      borderRadius: 'var(--radius-md)',
                      padding: '14px 16px',
                      fontSize: '0.85rem',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 8,
                    }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                        <span>Subtotal ({cartItems.reduce((acc, i) => acc + i.quantity, 0)} items)</span>
                        <span>${totalAmount.toFixed(2)}</span>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                        <span>Express Shipping</span>
                        <span style={{ color: 'var(--accent-emerald)', fontWeight: 600 }}>FREE</span>
                      </div>
                      <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        fontWeight: 700,
                        fontSize: '1.05rem',
                        color: 'var(--text-primary)',
                        paddingTop: 8,
                        borderTop: '1px solid var(--border-subtle)',
                      }}>
                        <span>Total</span>
                        <span style={{ color: 'var(--accent-cyan)' }}>${totalAmount.toFixed(2)}</span>
                      </div>
                    </div>
                  </form>
                </>
              )}
            </div>
          )}
        </div>

        <div className="modal-footer">
          {confirmedOrder ? (
            <button className="btn btn-primary" onClick={handleClose} style={{ width: '100%' }}>
              Continue Shopping
            </button>
          ) : (
            <>
              <button className="btn btn-ghost" onClick={handleClose} disabled={loading}>
                Continue Browsing
              </button>
              <button
                className="btn btn-emerald"
                onClick={handleCheckout}
                disabled={loading || cartItems.length === 0}
                id="confirm-checkout-btn"
              >
                {loading ? (
                  <>
                    <RefreshCw size={16} className="animate-spin" />
                    <span>Processing with Lock...</span>
                  </>
                ) : (
                  <>
                    <span>Pay ${totalAmount.toFixed(2)}</span>
                    <ArrowRight size={16} />
                  </>
                )}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
