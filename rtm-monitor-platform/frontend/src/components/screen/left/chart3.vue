<!-- 初始化代码 -->
<template>
  <div>
    <div>
      <!-- <span class="colorDeepskyblue" style="position: absolute; left: 10px;font-size: 10px;">时间：2022/07/01-2022/07/31</span>           -->
      <div class="colorGrass font-bold margin-l" :style="{'font-size': Math.round(YFOne*1.1) + 'px'}">
        路口车流量饱和度排行(%)
      </div>
    </div>
    <div>
      <div style="margin-left: 1.5%;">
        <div :style="{ height: YHOne + 'px',width: Math.round(YWOne*0.22) + 'px'}" ref="chartContainer"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref , onBeforeMount, onMounted, onUnmounted } from 'vue';
import * as echarts from 'echarts';

let screenHeight = ref(window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight)
let screenWidth = ref(window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth)

const chartContainer = ref<HTMLElement | null>(null);

// 浏览器高度
let YHOne = ref();
// 浏览器宽度
let YWOne = ref();
// 浏览器字体大小
let YFOne= ref();
// 自适应浏览器获取宽高大小定时器
const screenTimer = ref();

onBeforeMount( () => {
  YHOne.value = Math.round(screenHeight.value * 0.26);
})

onMounted( () => {
  // 页面大小改变时触发
  window.addEventListener('resize', getScreenHeight);
  // 页面大小改变时触发
  window.addEventListener('resize', getScreenWidth);
  // 鼠标移动时触发
  //window.addEventListener('mousemove',getHeight, false);
  // 自适应浏览器获取宽高大小定时器
  resizeScreen();
  // 获取接口数据
  getData();
  // 局部刷新定时器
  //getDataTimer();
})

onUnmounted( () => {
  // 清除自适应屏幕定时器
  clearInterval(screenTimer.value);
  screenTimer.value = null;
  // 页面大小改变时触发销毁
  window.removeEventListener('resize', getScreenHeight, false);
  // 页面大小改变时触发销毁
  window.removeEventListener('resize', getScreenWidth, false);
})

// 自适应浏览器获取宽高大小定时器
const resizeScreen = () => {
  screenTimer.value = setInterval(() => {
    getScreenHeight();
    getScreenWidth();
  }, 200)
}

// 获取浏览器高度进行自适应
const getScreenHeight = () => {
  screenHeight.value = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
  // 四舍五入取整数
  YHOne.value = Math.round(screenHeight.value * 0.26);
  //console.log("高度->"+screenHeight +"-"+ YHOne);
}
// 字体大小根据宽度自适应
const getScreenWidth = () => {
  screenWidth.value  = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
  // 浏览器字体计算
  YFOne.value  = Math.round(screenWidth.value  / 100);
  // 浏览器宽度宽度
  YWOne.value  = screenWidth.value;
  //console.log("宽度->"+screenWidth);
}

// 接口数据
const getData = () => {

  if (chartContainer.value) {
    const myChart = echarts.init(chartContainer.value);

    const option: echarts.EChartsOption = {
      xAxis: {
        type: 'category',
        data: ['秦直路','顺化路','安化路','庆州路','弘化路'],
      },
      yAxis: {
        type: 'value',
        name: '米',
      },
      color: ['#3398DB'],
      series: [
        {
          name: '实时',
          type: 'bar',
          data: [90, 86, 70, 65, 50],
        },
      ],
    };

    myChart.setOption(option);
  }
}

</script>
<style lang='scss' scoped>
.margin-l {
  margin-left: 25%;
}

// 字体颜色
// ::v-deep .dv-scroll-ranking-board .row-item {
//   color: aqua;
// }
</style>