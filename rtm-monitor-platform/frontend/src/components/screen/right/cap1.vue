<!-- 初始化代码 -->
<template>
  <div>
    <div>
      <!-- <span class="colorDeepskyblue" style="position: absolute; left: 10px;font-size: 10px;">时间：2022/07/01-2022/07/31</span>           -->
      <div class="colorGrass font-bold margin-l" :style="{'font-size': Math.round(YFOne*1.1) + 'px'}">
        岐黄大道-朔州东路
      </div>
    </div>
    <video id="my-video" class="video-js vjs-default-skin" autoplay muted preload="auto" controls  :style="{ height: Math.round(YHOne*0.90) + 'px',width: Math.round(YWOne*0.24) + 'px'}">
      <source src="http://124.152.91.184:20002/20003/video/index.m3u8"
          type="application/x-mpegURL" style='width: 100%;height: 100%'>
    </video>
  </div>
</template>

<script setup lang="ts">
import { ref, computed  , onBeforeMount, onMounted, onUnmounted,nextTick  } from 'vue';
import videojs from 'video.js/dist/video.min'
import 'video.js/dist/video-js.min.css'
// 获取浏览器可视区域高度（包含滚动条）、 window.innerHeight
// 获取浏览器可视区域高度（不包含工具栏高度）、document.documentElement.clientHeight
// 获取body的实际高度  (三个都是相同，兼容性不同的浏览器而设置的) document.body.clientHeight
let screenHeight = ref(window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight)
let screenWidth = ref(window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth)


// 浏览器高度
let YHOne = ref();
// 浏览器宽度
let YWOne = ref();
// 浏览器字体大小
let YFOne= ref();
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
  initWebRtcServer();
})

onUnmounted( () => {
  // 清除自适应屏幕定时器
  clearInterval(screenTimer.value);
  screenTimer.value = null;
  // 页面大小改变时触发销毁
  window.removeEventListener('resize', getScreenHeight, false);
  // 页面大小改变时触发销毁
  window.removeEventListener('resize', getScreenWidth, false);

  webRtcServerDis();
})

// 自适应浏览器获取宽高大小定时器
const resizeScreen = () => {
  screenTimer.value = setInterval(() => {
    getScreenHeight();
    getScreenWidth();
  }, 200)
}


const initWebRtcServer = async () => {
  videojs('my-video', {
    bigPlayButton: false,
    textTrackDisplay: false,
    posterImage: false,
    errorDisplay: false,
    controlBar: true,
    // ...其他配置参数
  }, function () {
    this.play()
  })
}
//页面销毁时销毁webRtc
const webRtcServerDis = () => {
  webRtcServer.disconnect()
  webRtcServer = null
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

</script>
<style lang='scss' scoped>
.video{
  margin-right: 2px;
}
.margin-l {
  margin-left: 35%;
}
</style>