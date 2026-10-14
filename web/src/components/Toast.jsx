import React from 'react';
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react';

export default function Toast({ toasts, onDismiss }) {
  if (!toasts || toasts.length === 0) return null;

  return (
    <div style={{
      position: 'fixed',
      bottom: 24,
      right: 24,
      display: 'flex',
      flexDirection: 'column',
      gap: 10,
      zIndex: 2000,
      pointerEvents: 'none',
    }}>
      {toasts.map((toast) => {
        const isSuccess = toast.type === 'success';
        const isError = toast.type === 'error';

        return (
          <div
            key={toast.id}
            style={{
              pointerEvents: 'auto',
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '12px 18px',
              borderRadius: 'var(--radius-md)',
              backgroundColor: 'var(--bg-secondary)',
              border: `1px solid ${
                isSuccess 
                  ? 'rgba(16, 185, 129, 0.4)' 
                  : isError 
                    ? 'rgba(244, 63, 94, 0.4)' 
                    : 'var(--border-subtle)'
              }`,
              boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
              maxWidth: 380,
              animation: 'slideUp 0.25s cubic-bezier(0.16, 1, 0.3, 1)',
            }}
          >
            {isSuccess && <CheckCircle2 size={18} style={{ color: 'var(--accent-emerald)', flexShrink: 0 }} />}
            {isError && <AlertCircle size={18} style={{ color: 'var(--accent-rose)', flexShrink: 0 }} />}
            {!isSuccess && !isError && <Info size={18} style={{ color: 'var(--accent-cyan)', flexShrink: 0 }} />}

            <div style={{ flex: 1, lineHeight: 1.3 }}>{toast.message}</div>

            <button
              onClick={() => onDismiss(toast.id)}
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text-muted)',
                cursor: 'pointer',
                padding: 4,
              }}
            >
              <X size={14} />
            </button>
          </div>
        );
      })}
    </div>
  );
}
