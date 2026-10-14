import React, { useState, useRef } from "react";
import {
  UploadCloud,
  FileText,
  CheckCircle2,
  AlertTriangle,
  X,
  Check,
  RefreshCw,
} from "lucide-react";
import { apiUrl } from "../config/api";

export default function CsvImportModal({ isOpen, onClose, onImportSuccess }) {
  const [file, setFile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [errorMsg, setErrorMsg] = useState("");
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef(null);

  if (!isOpen) return null;

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      setFile(e.dataTransfer.files[0]);
      setErrorMsg("");
      setResult(null);
    }
  };

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
      setErrorMsg("");
      setResult(null);
    }
  };

  const handleSubmit = async () => {
    if (!file) {
      setErrorMsg("Please select a CSV or TSV file to upload.");
      return;
    }

    setLoading(true);
    setErrorMsg("");
    setResult(null);

    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await fetch(apiUrl("/api/v1/products/import-csv"), {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Failed to import CSV file");
      }

      setResult(data);
      if (onImportSuccess) {
        onImportSuccess(data);
      }
    } catch (err) {
      setErrorMsg(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setFile(null);
    setResult(null);
    setErrorMsg("");
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className="modal-container"
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: result ? 780 : 580 }}
      >
        <div className="modal-header">
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <UploadCloud size={22} style={{ color: "var(--accent-cyan)" }} />
            <h2 style={{ fontSize: "1.2rem" }}>Bulk CSV Ingestion</h2>
          </div>
          <button
            className="btn btn-ghost"
            onClick={onClose}
            style={{ padding: 6, borderRadius: "50%" }}
          >
            <X size={18} />
          </button>
        </div>

        <div className="modal-body">
          {errorMsg && (
            <div
              style={{
                backgroundColor: "rgba(244, 63, 94, 0.12)",
                border: "1px solid rgba(244, 63, 94, 0.3)",
                color: "var(--accent-rose)",
                padding: "12px 16px",
                borderRadius: "var(--radius-md)",
                marginBottom: 16,
                fontSize: "0.85rem",
                display: "flex",
                alignItems: "center",
                gap: 8,
              }}
            >
              <AlertTriangle size={16} />
              <span>{errorMsg}</span>
            </div>
          )}

          {!result ? (
            <>
              <div
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                style={{
                  border: `2px dashed ${isDragging ? "var(--accent-cyan)" : "var(--border-subtle)"}`,
                  backgroundColor: isDragging
                    ? "var(--accent-cyan-glow)"
                    : "var(--bg-tertiary)",
                  borderRadius: "var(--radius-lg)",
                  padding: "36px 20px",
                  textAlign: "center",
                  cursor: "pointer",
                  transition: "var(--transition)",
                  marginBottom: 16,
                }}
              >
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  accept=".csv,.tsv,.txt"
                  style={{ display: "none" }}
                />

                <div
                  style={{
                    width: 52,
                    height: 52,
                    borderRadius: "50%",
                    backgroundColor: "rgba(6, 182, 212, 0.1)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    margin: "0 auto 12px",
                  }}
                >
                  <UploadCloud
                    size={28}
                    style={{ color: "var(--accent-cyan)" }}
                  />
                </div>

                <div
                  style={{
                    fontWeight: 600,
                    fontSize: "0.95rem",
                    marginBottom: 4,
                  }}
                >
                  {file ? file.name : "Choose a CSV file or drag & drop here"}
                </div>

                {file && (
                  <div
                    style={{
                      marginTop: 12,
                      display: "inline-flex",
                      alignItems: "center",
                      gap: 6,
                      fontSize: "0.8rem",
                      color: "var(--accent-cyan)",
                    }}
                  >
                    <FileText size={14} />
                    <span>{(file.size / 1024).toFixed(1)} KB selected</span>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div>
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "repeat(4, 1fr)",
                  gap: 12,
                  marginBottom: 20,
                }}
              >
                <div
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    padding: "14px",
                    borderRadius: "var(--radius-md)",
                    textAlign: "center",
                  }}
                >
                  <div
                    style={{ fontSize: "0.7rem", color: "var(--text-muted)" }}
                  >
                    TOTAL ROWS
                  </div>
                  <div style={{ fontSize: "1.4rem", fontWeight: 700 }}>
                    {result.total_rows}
                  </div>
                </div>

                <div
                  style={{
                    backgroundColor: "rgba(16, 185, 129, 0.1)",
                    border: "1px solid rgba(16, 185, 129, 0.25)",
                    padding: "14px",
                    borderRadius: "var(--radius-md)",
                    textAlign: "center",
                  }}
                >
                  <div
                    style={{
                      fontSize: "0.7rem",
                      color: "var(--accent-emerald)",
                    }}
                  >
                    NEW CREATED
                  </div>
                  <div
                    style={{
                      fontSize: "1.4rem",
                      fontWeight: 700,
                      color: "var(--accent-emerald)",
                    }}
                  >
                    {result.created}
                  </div>
                </div>

                <div
                  style={{
                    backgroundColor: "rgba(6, 182, 212, 0.1)",
                    border: "1px solid rgba(6, 182, 212, 0.25)",
                    padding: "14px",
                    borderRadius: "var(--radius-md)",
                    textAlign: "center",
                  }}
                >
                  <div
                    style={{ fontSize: "0.7rem", color: "var(--accent-cyan)" }}
                  >
                    UPDATED
                  </div>
                  <div
                    style={{
                      fontSize: "1.4rem",
                      fontWeight: 700,
                      color: "var(--accent-cyan)",
                    }}
                  >
                    {result.updated}
                  </div>
                </div>

                <div
                  style={{
                    backgroundColor:
                      result.failed > 0
                        ? "rgba(244, 63, 94, 0.1)"
                        : "var(--bg-tertiary)",
                    border:
                      result.failed > 0
                        ? "1px solid rgba(244, 63, 94, 0.25)"
                        : "none",
                    padding: "14px",
                    borderRadius: "var(--radius-md)",
                    textAlign: "center",
                  }}
                >
                  <div
                    style={{
                      fontSize: "0.7rem",
                      color:
                        result.failed > 0
                          ? "var(--accent-rose)"
                          : "var(--text-muted)",
                    }}
                  >
                    REJECTED
                  </div>
                  <div
                    style={{
                      fontSize: "1.4rem",
                      fontWeight: 700,
                      color:
                        result.failed > 0
                          ? "var(--accent-rose)"
                          : "var(--text-primary)",
                    }}
                  >
                    {result.failed}
                  </div>
                </div>
              </div>

              {result.errors && result.errors.length > 0 ? (
                <div>
                  <div
                    style={{
                      fontSize: "0.85rem",
                      fontWeight: 600,
                      marginBottom: 8,
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                      color: "var(--accent-rose)",
                    }}
                  >
                    <AlertTriangle size={16} />
                    <span>
                      Row-Level Rejection Audit ({result.errors.length} errors)
                    </span>
                  </div>

                  <div
                    style={{
                      maxHeight: 260,
                      overflowY: "auto",
                      border: "1px solid var(--border-subtle)",
                      borderRadius: "var(--radius-md)",
                      backgroundColor: "var(--bg-tertiary)",
                    }}
                  >
                    <table
                      style={{
                        width: "100%",
                        borderCollapse: "collapse",
                        fontSize: "0.75rem",
                        textAlign: "left",
                      }}
                    >
                      <thead>
                        <tr
                          style={{
                            backgroundColor: "rgba(0, 0, 0, 0.25)",
                            borderBottom: "1px solid var(--border-subtle)",
                            position: "sticky",
                            top: 0,
                          }}
                        >
                          <th style={{ padding: "8px 12px", width: 60 }}>
                            Row
                          </th>
                          <th style={{ padding: "8px 12px", width: 100 }}>
                            SKU
                          </th>
                          <th style={{ padding: "8px 12px", width: 100 }}>
                            Field
                          </th>
                          <th style={{ padding: "8px 12px" }}>Reason</th>
                        </tr>
                      </thead>
                      <tbody>
                        {result.errors.map((err, idx) => (
                          <tr
                            key={idx}
                            style={{
                              borderBottom:
                                "1px solid rgba(255, 255, 255, 0.04)",
                            }}
                          >
                            <td
                              style={{
                                padding: "8px 12px",
                                color: "var(--text-muted)",
                              }}
                            >
                              #{err.row}
                            </td>
                            <td style={{ padding: "8px 12px" }}>
                              {err.sku ? (
                                <code>{err.sku}</code>
                              ) : (
                                <span style={{ color: "var(--text-muted)" }}>
                                  -
                                </span>
                              )}
                            </td>
                            <td
                              style={{
                                padding: "8px 12px",
                                color: "var(--accent-amber)",
                              }}
                            >
                              {err.field}
                            </td>
                            <td
                              style={{
                                padding: "8px 12px",
                                color: "var(--text-secondary)",
                              }}
                            >
                              {err.error}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <div
                  style={{
                    padding: "20px",
                    textAlign: "center",
                    backgroundColor: "rgba(16, 185, 129, 0.08)",
                    borderRadius: "var(--radius-md)",
                    color: "var(--accent-emerald)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    gap: 8,
                    fontSize: "0.9rem",
                  }}
                >
                  <CheckCircle2 size={20} />
                  <span>All rows imported with 100% compliance!</span>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="modal-footer">
          {!result ? (
            <>
              <button
                className="btn btn-ghost"
                onClick={onClose}
                disabled={loading}
              >
                Cancel
              </button>
              <button
                className="btn btn-primary"
                onClick={handleSubmit}
                disabled={!file || loading}
                id="submit-csv-btn"
              >
                {loading ? (
                  <>
                    <RefreshCw size={16} className="animate-spin" />
                    <span>Processing...</span>
                  </>
                ) : (
                  <>
                    <UploadCloud size={16} />
                    <span>Start Ingestion</span>
                  </>
                )}
              </button>
            </>
          ) : (
            <>
              <button className="btn btn-secondary" onClick={handleReset}>
                Import Another File
              </button>
              <button className="btn btn-emerald" onClick={onClose}>
                <Check size={16} />
                <span>Done</span>
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
