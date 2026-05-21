<template>
  <div ref="mapContainer" style="height: 300px;"></div>
</template>
<script lang='js' setup>
import { ref, reactive, onMounted, nextTick } from 'vue'

import "ol/ol.css";
import { Map, View, Feature, Overlay } from "ol";
import { Vector as VectorLayer } from "ol/layer";
import { Vector as VectorSource } from "ol/source";
import { Point, LineString } from "ol/geom.js";
import { Icon, Fill, Stroke, Style, Circle } from "ol/style";
import { getVectorContext } from "ol/render";
import { ElMessage } from 'element-plus';
import startIcon from '@/assets/images/start_point.png'
import endIcon from '@/assets/images/end_point.png'
import walkIcon from '@/assets/images/walk-icon.png'
import rightIcon from '@/assets/images/right-icon.png'

const parentData = defineProps(["map", "view"]);
const map = parentData.map;
const view = parentData.view;
const mapContainer = ref(null);

let geometryMove = {}
let featureMove = {}
let vectorLayer = ref(null)

let distance = 0
let lastTime = 0
let speed = 0.1

let styles = {
  route: new Style({
    stroke: new Stroke({
      width: 8,
      color: "green",
    }),
  }),
  iconStart: new Style({
    image: new Icon({
      anchor: [0.5, 1],
      src: startIcon,
      scale: 0.2, //设置大小
    }),
  }),
  iconEnd: new Style({
    image: new Icon({
      anchor: [0.5, 1],
      src: endIcon,
      scale: 0.2, //设置大小
    }),
  }),
  featureMove: new Style({
    image: new Icon({
      anchor: [0.5, 1],
      src: walkIcon,
      scale: 0.2, //设置大小
    }),
  }),
}

let route = null

const drawHandle = (coordinateArr) => {
  clearDrawHandle()
  view.setCenter(coordinateArr[0])
  route = new LineString(coordinateArr)
  geometryMove = new Point(route.getFirstCoordinate())
  featureMove = new Feature({
    type: "featureMove",
    geometry: geometryMove,
  })
  vectorLayer.value = new VectorLayer({
    source: new VectorSource({
      features: [
        new Feature({
          type: "route",
          geometry: route,
        }),
        featureMove,
        new Feature({
          type: "iconStart",
          geometry: new Point(route.getFirstCoordinate()),
        }),
        new Feature({
          type: "iconEnd",
          geometry: new Point(route.getLastCoordinate()),
        }),
      ],
    }),
    style: (feature) => {
      if (feature.get("type") == 'route') {
        feature.setStyle(arrowLineStyles)
        return
      }

      return styles[feature.get("type")];
    },
  })

  map.addLayer(vectorLayer.value)
}

// 清除绘制
const clearDrawHandle = () => {
  if (vectorLayer.value) {
    // vectorLayer.value.getSource().clear()
    map.removeLayer(vectorLayer.value)
    vectorLayer.value = null
  }
}

// 移动
const moveFeature = (e) => {
  let time = e.frameState.time;
  distance =
    (distance + (speed * (time - lastTime)) / 2000) % 1; //%2表示：起止止起；%1表示：起止起止

  lastTime = time;

  const currentCoordinate = route.getCoordinateAt(
    distance > 1 ? 2 - distance : distance
  );
  geometryMove.setCoordinates(currentCoordinate);
  const vectorContext = getVectorContext(e);


  vectorContext.setStyle(styles.featureMove);
  vectorContext.drawGeometry(geometryMove);
  map.render();
}

// 动画开始
const startAnimation = () => {
  if (vectorLayer.value) {
    lastTime = Date.now();
    vectorLayer.value.on("postrender", moveFeature);
    featureMove.setGeometry(null); //必须用null，不能用{}
  } else {
    ElMessage.warning('请先绘制路线！')
  }
}

// 动画结束
const stopAnimation = () => {
  featureMove.setGeometry(geometryMove);
  vectorLayer.value.un("postrender", moveFeature);
}


// 箭头样式
const arrowLineStyles = (feature, resolution) => {
  let styles = [];
  // 线条样式
  let backgroundLineStyle = new Style({
    stroke: new Stroke({
      width: 10,
      color: "green",
    }),
  });
  styles.push(backgroundLineStyle);
  let geometry = feature.getGeometry();
  // 获取线段长度
  const length = geometry.getLength();
  // 箭头间隔距离（像素）
  const step = 50;
  // 将间隔像素距离转换成地图的真实距离
  const StepLength = step * resolution;
  // 得到一共需要绘制多少个 箭头
  const arrowNum = Math.floor(length / StepLength);
  const rotations = [];
  const distances = [0];
  geometry.forEachSegment(function (start, end) {
    let dx = end[0] - start[0];
    let dy = end[1] - start[1];
    let rotation = Math.atan2(dy, dx);
    distances.unshift(Math.sqrt(dx ** 2 + dy ** 2) + distances[0]);
    rotations.push(rotation);
  });
  // 利用之前计算得到的线段矢量信息，生成对应的点样式塞入默认样式中
  // 从而绘制内部箭头
  for (let i = 1; i < arrowNum; ++i) {
    const arrowCoord = geometry.getCoordinateAt(i / arrowNum);
    const d = i * StepLength;
    const grid = distances.findIndex((x) => x <= d);

    styles.push(
      new Style({
        geometry: new Point(arrowCoord),
        image: new Icon({
          src: rightIcon,
          opacity: 1,
          anchor: [0.5, 0.5],
          rotateWithView: false,
          // 读取 rotations 中计算存放的方向信息
          rotation: -rotations[distances.length - grid - 1],
          scale: 0.1,
        }),
      })
    );
  }
  return styles;
}

defineExpose({
  drawHandle,
  clearDrawHandle,
  startAnimation,
  stopAnimation
})


</script>
<style scoped lang='less'></style>