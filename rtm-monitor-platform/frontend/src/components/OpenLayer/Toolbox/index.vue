<template>
  <div class="toolbox">
    <!-- 工具箱的按钮 -->
    <!-- ... 其他按钮 ... -->
    <button @click="onToolboxAction('zoomIn')">放大</button>
    <button @click="onToolboxAction('zoomOut')">缩小</button>
    <!-- ... 其他按钮 ... -->
    <!-- 更多按钮 -->
  </div>
</template>

<script setup>
import { defineEmits } from 'vue'
// 定义emit
const emit = defineEmits(['toolboxAction'])

function onToolboxAction(action) {
  // 发出事件，action是操作的名称，如'zoomIn'或'zoomOut'
  emit('toolboxAction', action)
}
</script>

<style scoped>
.toolbox {
  position: absolute;
  top: 10%; /* 垂直居中 */
  transform: translateY(-50%); /* 确保工具箱垂直居中 */
  right: 0; /* 紧贴右边缘 */
  z-index: 1000; /* 确保工具箱在地图上方 */
  background: #20498E; /* 工具箱背景颜色 */
  border-radius: 5px; /* 工具箱圆角 */
  display: flex;
  flex-direction: column; /* 垂直排列按钮 */
  padding: 2px;
}

.tool-btn {
  background: #555; /* 按钮背景颜色 */
  color: white; /* 按钮文字颜色 */
  border: none; /* 移除边框 */
  margin: 5px 0; /* 增加按钮之间的间隙 */
  padding: 2px;
  cursor: pointer; /* 鼠标悬停时显示指针 */
  text-align: center; /* 文字居中 */
  border-radius: 3px; /* 按钮圆角 */
}

.tool-btn:hover {
  background: #666; /* 鼠标悬停时按钮背景颜色 */
}
</style>
