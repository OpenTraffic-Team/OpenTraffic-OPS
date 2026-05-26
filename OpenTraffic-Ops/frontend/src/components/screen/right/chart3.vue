<!-- 初始化代码 -->
<template>
  <div>
    <div>
      <div class="colorGrass font-bold margin-l" :style="{'font-size': Math.round(YFOne*1.1) + 'px'}">
        空放时长排名
      </div>
    </div>
    <div>
      <dv-capsule-chart :config="config" :style="{ height: YHOne + 'px',width: Math.round(YWOne*0.20) + 'px'}" />
    </div>
  </div>    
  
</template>

<script setup lang="ts">
import { ref, reactive , onBeforeMount, onMounted, onUnmounted } from 'vue';
  // 获取浏览器可视区域高度（包含滚动条）、 window.innerHeight
  // 获取浏览器可视区域高度（不包含工具栏高度）、document.documentElement.clientHeight
  // 获取body的实际高度  (三个都是相同，兼容性不同的浏览器而设置的) document.body.clientHeight
  let screenHeight = ref(window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight)
  let screenWidth = ref(window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth)
  const config = reactive({
    data: [
      {
        name: '石油路',
        value: 120,
      },
      {
        name: '正宁路',
        value: 90,
      },
      {
        name: '宁县路',
        value: 78,
      },
      {
        name: '庆城路',
        value: 66,
      },
      {
        name: '合水路',
        value: 60,
      },
      {
        name: '华池路',
        value: 45,
      }
    ],
    // unit: '万元',
    showValue: true
 })

  // 浏览器高度
  let YHOne = ref();
  // 浏览器宽度
  let YWOne = ref();
  // 浏览器字体大小
  let YFOne= ref();
  // 自适应浏览器获取宽高大小定时器
  const screenTimer = ref(); 

onBeforeMount( () => {
  YHOne.value = Math.round(screenHeight.value * 0.20);
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
      YHOne.value = Math.round(screenHeight.value * 0.3);
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
    // 调用接口方法
    // getModuleData().then(res => {
    //       allData = res.data;
    //       //console.log("ALLDATA->"+JSON.stringify(allData.pcsSum));
    //       // echarts查不到数据，将初始化echarts的方法全部放置到接口方法中即可。  
    // })  
    // 获取服务器的数据, 对allData进行赋值之后, 调用updateChart方法更新图表
    //console.log("ALLDATA->",JSON.stringify(res.data))
    //console.log("ALLDATA->",JSON.stringify(allData))
  }
  // 定时器
  const getDataTimer = () => {
    screenTimer.value = setInterval(() => {
      // 执行刷新数据的方法
      getData();
      //console.log("Hello World")
    }, 3000)
  }
</script>
<style lang='scss' scoped>
.margin-l {
  margin-left: 35%;
}

// 让下方单位的精度值隐藏
::v-deep(.dv-capsule-chart .unit-label)  {
  visibility: hidden;
}
</style>