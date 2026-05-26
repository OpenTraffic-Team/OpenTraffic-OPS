<script setup lang="ts">
import {onBeforeMount, onMounted, onUnmounted, ref} from 'vue';
import Map from 'ol/Map';
import * as olProj from 'ol/proj'
import { Tile as TileLayer } from 'ol/layer';
import VectorLayer from "ol/layer/Vector";
import VectorSource from "ol/source/Vector";
import Feature from 'ol/Feature';
import View from 'ol/View';
import  XYZ from "ol/source/XYZ";
import { Style, Icon } from "ol/style";
import Point from 'ol/geom/Point';
import trackCar from "../../../assets/map/light.json";
import lightPng from '../../../assets/images/ai-light-green.png';


const mapContainer = ref<HTMLElement | null>(null);

let screenHeight = ref(window.innerHeight || document.documentElement.clientHeight ||document.body.clientHeight);
let screenWidth = ref(window.innerWidth || document.documentElement.clientWidth ||document.body.clientWidth);

const screenTimer = ref();
const dataTimer = ref();

let YHOne = ref();
let YFOne = ref();

let map=null;
let vectorLayer=null;

onBeforeMount( () => {
  YHOne.value = Math.round(screenHeight.value * 0.90);
  YFOne.value = Math.round(screenWidth.value * 0.49);

})

onMounted(()=>{
  addTrack();
  initMap();
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
  YHOne.value = Math.round(screenHeight.value * 0.90);
  //console.log("高度->"+screenHeight +"-"+ kHOne);
}
// 字体大小根据宽度自适应
const getScreenWidth = () => {
  screenWidth.value = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
  YFOne.value = Math.round(screenWidth.value * 0.49);
  //console.log("宽度->"+screenWidth);
}
const addTrack = () => {
  // 创建开始图标
  const startMarker = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[0])),
  });

  const marker1 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[1])),
  });
  const marker2 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[2])),
  });
  const marker3 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[3])),
  });
  const marker4 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[4])),
  });
  const marker5 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[5])),
  });
  const marker6 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[6])),
  });
  const marker7 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[7])),
  });
  const marker8 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[8])),
  });
  const marker9 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[9])),
  });
  const marker10 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[10])),
  });
  const marker11 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[11])),
  });
  const marker12 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[12])),
  });
  // 创建结束图标
  const endMarker = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat(trackCar[trackCar.length - 1])),
  });


  const estMarker1 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat([107.65915453,35.69562170])),
  });
  const estMarker2 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat([107.65963733,35.69311226])),
  });
  const estMarker3 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat([107.66046345,35.68876848])),
  });
  const estMarker4 = new Feature({
    type: "marker",
    geometry: new Point(olProj.fromLonLat([107.66142368,35.68351381])),
  });


  // 设置样式
  var styles = {
    // 如果类型是 geoMarker 的样式
    marker: new Style({
      image: new Icon({
        src: lightPng,
        anchor: [0.5, 1.1], // 设置偏移
        scale: 0.3
      }),
    }),
  };
  // 把小车和线添加到图层
  vectorLayer = new VectorLayer({
    source: new VectorSource({
      features: [startMarker,marker1,marker2,marker3,marker4,marker5,marker6, marker7,marker8,
        marker9,marker10,marker11,marker12,endMarker,estMarker1,estMarker2,estMarker3,estMarker4],
    }), //线,起点的图标,终点的图标
    style: function (feature) {
      return styles[feature.get("type")];
    },
  });
};

const initMap = () => {

  if (mapContainer.value) {

    map = new Map({
      target: mapContainer.value,
      layers: [
        new TileLayer({
          source:new XYZ({
            url: 'https://wprd0{1-4}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&style=7&x={x}&y={y}&z={z}',
            crossOrigin: "anonymous",
            tileLoadFunction:  (imageTile, src)=> {
              // 使用滤镜 将白色修改为深色
              let img = new Image()
              // img.crossOrigin = ''
              // 设置图片不从缓存取，从缓存取可能会出现跨域，导致加载失败
              img.setAttribute('crossOrigin', 'anonymous')
              img.onload =  ()=> {
                let canvas = document.createElement('canvas')
                let w = img.width
                let h = img.height
                canvas.width = w
                canvas.height = h
                let context = canvas.getContext('2d')
                context.filter = 'grayscale(98%) invert(100%) sepia(20%) hue-rotate(180deg) saturate(1600%) brightness(80%) contrast(90%)'
                context.drawImage(img, 0, 0, w, h, 0, 0, w, h)
                imageTile.getImage().src = canvas.toDataURL('image/png')
              },
                  img.onerror = ()=>{
                    imageTile.getImage().src = require('@/assets/404_images/404.png')
                  }
              img.src = src
            },
          })
        }),
        vectorLayer
      ],
      view: new View({
        center: olProj.fromLonLat([107.65,35.71]),
        zoom : 14,
        minZoom : 0,
        maxZoom : 16
      })
    });
    map.getView().on('change:resolution', () => {
      const zoom = map.getView().getZoom();
      updateMarkerSize(zoom);
    });
  }else {
    console.log(mapContainer)
  }
};
// 更新标记的大小
const updateMarkerSize = (zoom) => {
  const scaleFactor = calculateScaleFactor(zoom); // 计算缩放因子

  // 更新起点和终点标记的样式
  vectorLayer.getSource().getFeatures().forEach(feature => {
    if (feature.get('type') === 'startMarker' || feature.get('type') === 'endMarker') {
      const style = feature.getStyle();
      if (style && style.getImage()) { // 确保样式和图标存在
        const iconStyle = style.getImage();
        iconStyle.setScale(scaleFactor); // 设置新的缩放级别
        feature.setStyle(new Style({ image: iconStyle }));
      }
    }
  });
};

// 根据缩放级别计算缩放因子
const calculateScaleFactor = (zoom) => {
  // 示例：根据缩放级别计算缩放因子
  // 这里您可以根据需要调整计算方式
  return 0.1 * zoom; // 示例缩放因子
};


</script>

<template>
  <div class="home">
    <div style="width: 100%; height: 100%">
      <div  ref="mapContainer"  :style="{ height: YHOne + 'px',width: YFOne + 'px'}"></div>
    </div>
  </div>
</template>

<style scoped>
</style>
