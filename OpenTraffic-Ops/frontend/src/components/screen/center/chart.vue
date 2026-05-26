<template>
  <div id="center">
    <div style="margin-right: 2px;">
      <img src="@/assets/images/map.png"   :style="{ height: YHOne + 'px',width: YFOne + 'px'}">
    </div>
  </div>
</template>

<script setup lang="ts">

import { ref, onBeforeMount, onMounted, onUnmounted } from 'vue';
    // 获取浏览器可视区域高度（包含滚动条）、 window.innerHeight
    // 获取浏览器可视区域高度（不包含工具栏高度）、document.documentElement.clientHeight
    // 获取body的实际高度  (三个都是相同，兼容性不同的浏览器而设置的) document.body.clientHeight
    let screenHeight = ref(window.innerHeight || document.documentElement.clientHeight ||document.body.clientHeight);
    let screenWidth = ref(window.innerWidth || document.documentElement.clientWidth ||document.body.clientWidth);
    const screenTimer = ref();
    const dataTimer = ref();

    let YHOne = ref();
    let YFOne = ref();

  onBeforeMount( () => {
    YHOne.value = Math.round(screenHeight.value * 0.92);
    YFOne.value = Math.round(screenWidth.value * 0.49);

  })  

  onMounted(() => {
    // 页面大小改变时触发
    window.addEventListener("resize", getScreenHeight, false);
    // 页面大小改变时触发
    window.addEventListener("resize", getScreenWidth, false);
    // 鼠标移动时触发
    // window.addEventListener('mousemove',getHeight, false);
    resizeScreen();
  })

  onUnmounted( () => {
    // 清除多次执行定时器
    clearInterval(screenTimer.value);
    screenTimer.value = null;
    // 清除多次执行定时器
    clearInterval(dataTimer.value);
    dataTimer.value = null;
    // 页面大小改变时触发
    window.removeEventListener("resize", getScreenHeight, false);
    // 页面大小改变时触发
    window.removeEventListener("resize", getScreenWidth, false);
  }) 
  
  // 自适应监控定时器
  const resizeScreen = () => {
      screenTimer.value = setInterval(() => {
        getScreenHeight();
        getScreenWidth();
      }, 200);
    }

  // 获取浏览器高度进行自适应
  const getScreenHeight = () => {
      screenHeight.value = window.innerHeight || document.documentElement.clientHeight ||document.body.clientHeight
      // 四舍五入取整数
      YHOne.value = Math.round(screenHeight.value * 0.92);
      //console.log("高度->"+screenHeight +"-"+ kHOne);
    }
    // 字体大小根据宽度自适应
  const getScreenWidth = () => {
      screenWidth.value = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
      YFOne.value = Math.round(screenWidth.value * 0.49);
      //console.log("宽度->"+screenWidth);
    }
</script>

<style lang="scss" scoped>
#center {
  display: flex;
  flex-direction: column;
  .square {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    justify-content: space-around;
    .item {
      // 控制方块宽度比例
      width: 24.5%;
      border-radius: 6px;
      margin-top: 0.5%;
      margin-bottom: 0.5%;
    }
  }
}
</style>
