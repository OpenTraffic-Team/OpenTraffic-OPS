<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: 'Microsoft YaHei', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100vh;
  overflow: hidden;
}

html, body {
  height: 100%;
  overflow: hidden;
}

/* ========== 全局对话框覆盖 ========== */
body .dark-dialog,
body .dark-dialog.el-overlay {
  --el-dialog-bg-color: #ffffff !important;
  --el-text-color-regular: #374151 !important;
}

body .dark-dialog .el-dialog,
.dark-dialog.el-overlay .el-dialog {
  background-color: #ffffff !important;
  border: 1px solid #e5e7eb !important;
  border-radius: 16px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15) !important;
}

body .dark-dialog .el-dialog__header,
.dark-dialog.el-overlay .el-dialog__header {
  padding: 20px 24px !important;
  border-bottom: 1px solid #f3f4f6 !important;
  margin-right: 0 !important;
  background-color: transparent !important;
}

body .dark-dialog .el-dialog__title,
.dark-dialog.el-overlay .el-dialog__title {
  color: #1f2937 !important;
  font-size: 16px !important;
  font-weight: 600 !important;
}

body .dark-dialog .el-dialog__headerbtn .el-dialog__close,
.dark-dialog.el-overlay .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af !important;
}

body .dark-dialog .el-dialog__headerbtn:hover .el-dialog__close,
.dark-dialog.el-overlay .el-dialog__headerbtn:hover .el-dialog__close {
  color: #6b7280 !important;
}

body .dark-dialog .el-dialog__body,
.dark-dialog.el-overlay .el-dialog__body {
  background-color: #ffffff !important;
  color: #374151 !important;
  padding: 24px !important;
}

body .dark-dialog .el-dialog__footer,
.dark-dialog.el-overlay .el-dialog__footer {
  background-color: #ffffff !important;
  padding: 16px 24px !important;
  border-top: 1px solid #f3f4f6 !important;
}

/* 表单 */
body .dark-form .el-form-item__label,
.dark-dialog .dark-form .el-form-item__label {
  color: #6b7280 !important;
  font-weight: 500 !important;
}

body .dark-form .el-input__wrapper,
.dark-dialog .dark-form .el-input__wrapper {
  background-color: #f9fafb !important;
  border: 1px solid #e5e7eb !important;
  box-shadow: none !important;
  border-radius: 10px !important;
}

body .dark-form .el-input__inner,
.dark-dialog .dark-form .el-input__inner {
  color: #374151 !important;
}

body .dark-form .el-input__inner::placeholder,
.dark-dialog .dark-form .el-input__inner::placeholder {
  color: #9ca3af !important;
}

body .dark-form .el-input__wrapper:hover,
.dark-dialog .dark-form .el-input__wrapper:hover {
  border-color: #d1d5db !important;
}

body .dark-form .el-input__wrapper.is-focus,
.dark-dialog .dark-form .el-input__wrapper.is-focus {
  border-color: rgba(99, 102, 241, 0.5) !important;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1) !important;
}

/* 对话框按钮 */
body .dark-dialog .el-button:not(.el-button--primary),
.dark-dialog.el-overlay .el-button:not(.el-button--primary) {
  background-color: #f3f4f6 !important;
  border: 1px solid #e5e7eb !important;
  color: #6b7280 !important;
}

body .dark-dialog .el-button:not(.el-button--primary):hover,
.dark-dialog.el-overlay .el-button:not(.el-button--primary):hover {
  background-color: #e5e7eb !important;
  border-color: #d1d5db !important;
  color: #374151 !important;
}

body .dark-dialog .el-button--primary,
.dark-dialog.el-overlay .el-button--primary {
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%) !important;
  border: none !important;
}

body .dark-dialog .el-button--primary:hover,
.dark-dialog.el-overlay .el-button--primary:hover {
  background: linear-gradient(135deg, #818cf8 0%, #a78bfa 100%) !important;
}

/* ========== 全局表格覆盖 ========== */
body .dark-table,
body .dark-table.el-table {
  --el-table-bg-color: transparent !important;
  --el-table-tr-bg-color: transparent !important;
  --el-table-header-bg-color: #f9fafb !important;
  --el-table-border-color: #f3f4f6 !important;
  --el-table-row-hover-bg-color: #f9fafb !important;
  --el-table-text-color: #374151 !important;
  --el-table-header-text-color: #6b7280 !important;
}

/* 隐藏表格所有装饰性伪元素 */
body .dark-table .el-table__inner-wrapper::before,
body .dark-table .el-table__inner-wrapper::after,
.dark-table.el-table .el-table__inner-wrapper::before,
.dark-table.el-table .el-table__inner-wrapper::after {
  display: none !important;
}

body .dark-table .el-table__header-wrapper::after,
.dark-table.el-table .el-table__header-wrapper::after {
  display: none !important;
}

/* 强制覆盖 header 单元格样式 */
body .dark-table .el-table__header th.el-table__cell,
.dark-table.el-table .el-table__header th.el-table__cell {
  background-color: #f9fafb !important;
  color: #6b7280 !important;
  border-bottom-color: #f3f4f6 !important;
}

/* 强制覆盖 body 单元格样式 */
body .dark-table .el-table__row .el-table__cell,
.dark-table.el-table .el-table__row .el-table__cell {
  background-color: transparent !important;
  color: #374151 !important;
  border-bottom-color: #f3f4f6 !important;
}

body .dark-table .el-table__row:hover .el-table__cell,
.dark-table.el-table .el-table__row:hover .el-table__cell {
  background-color: #f9fafb !important;
}

body .dark-table .el-table__empty-text,
.dark-table.el-table .el-table__empty-text {
  color: #9ca3af !important;
}

/* Alert */
.dark-alert.el-alert--info {
  background: rgba(99, 102, 241, 0.08) !important;
  border: 1px solid rgba(99, 102, 241, 0.15) !important;
  border-radius: 10px !important;
}

.dark-alert.el-alert--info .el-alert__title {
  color: #6366f1 !important;
  font-size: 13px;
}

.dark-alert.el-alert--info .el-alert__description {
  color: rgba(99, 102, 241, 0.7) !important;
  font-size: 13px;
}

.dark-alert .el-alert__icon {
  color: #6366f1 !important;
}

/* 弹窗确认框 */
.el-message-box {
  background: #ffffff !important;
  border: 1px solid #e5e7eb !important;
  border-radius: 16px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15) !important;
}

.el-message-box__title {
  color: #1f2937 !important;
}

.el-message-box__message {
  color: #6b7280 !important;
}

.el-message-box__headerbtn .el-message-box__close {
  color: #9ca3af !important;
}

.el-message-box__headerbtn:hover .el-message-box__close {
  color: #6b7280 !important;
}

/* 消息提示 */
.el-message {
  background: #ffffff !important;
  border: 1px solid #e5e7eb !important;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.15) !important;
}

.el-message .el-message__content {
  color: #374151 !important;
}

/* 分页 */
body .dark-table .el-pagination,
.dark-table.el-table .el-pagination {
  color: #6b7280;
}

body .dark-table .el-pagination .btn-prev,
body .dark-table .el-pagination .btn-next,
.dark-table.el-table .el-pagination .btn-prev,
.dark-table.el-table .el-pagination .btn-next {
  background: #f3f4f6;
  color: #6b7280;
}

body .dark-table .el-pagination .el-pager li,
.dark-table.el-table .el-pagination .el-pager li {
  background: #f3f4f6;
  color: #6b7280;
}

body .dark-table .el-pagination .el-pager li.is-active,
.dark-table.el-table .el-pagination .el-pager li.is-active {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
}
</style>
