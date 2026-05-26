<template>
  <!-- 设置span中的内容左对齐-->
  <span>
    <template v-for="(item, index) in values">
      <span>
         <!-- 设置图片的宽和高 -->
         <el-image v-if="item === 'EW'" :src="EW" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'NE'" :src="NE" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'NS'" :src="NS" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'SN'" :src="SN" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'WE'" :src="WE" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'SW'" :src="SW" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'WN'" :src="WN" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'ES'" :src="ES" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'EW_ES'" :src="EW_ES" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'NS_NE'" :src="NS_NE" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'SN_SW'" :src="SN_SW" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'WE_WN'" :src="WE_WN" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'SW_SS'" :src="SW_SS" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'WN_WW'" :src="WN_WW" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'NE_NN'" :src="NE_NN" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'ES_EE'" :src="ES_EE" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'SS'" :src="SS" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'WW'" :src="WW" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'NN'" :src="NN" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <el-image v-else-if="item === 'EE'" :src="EE" :style="{ width: props.imageWidth, height: props.imageHeight }" style="margin: 2px"/>
         <span v-else>{{ item }}</span>
      </span>
    </template>
    <template v-if="showValue">
      {{ values }}
    </template>
  </span>
</template>

<script setup>
import EN from '@/assets/images/phase/EN.png';
import ES from '@/assets/images/phase/ES.png';
import EW from '@/assets/images/phase/EW.png';
import EW_ES from '@/assets/images/phase/EW_ES.png';
import NE from '@/assets/images/phase/NE.png';
import NS from '@/assets/images/phase/NS.png';
import NS_NE from '@/assets/images/phase/NS_NE.png';
import NW from '@/assets/images/phase/NW.png';
import SE from '@/assets/images/phase/SE.png';
import SN from '@/assets/images/phase/SN.png';
import SN_SW from '@/assets/images/phase/SN_SW.png';
import SW from '@/assets/images/phase/SW.png';
import WE from '@/assets/images/phase/WE.png';
import WE_WN from '@/assets/images/phase/WE_WN.png';
import WN from '@/assets/images/phase/WN.png';
import WS from '@/assets/images/phase/WS.png';

import SW_SS from '@/assets/images/phase/SW_SS.png';
import WN_WW from '@/assets/images/phase/WN_WW.png';
import NE_NN from '@/assets/images/phase/NE_NN.png';
import ES_EE from '@/assets/images/phase/ES_EE.png';
import SS from '@/assets/images/phase/SS.png';
import WW from '@/assets/images/phase/WW.png';
import NN from '@/assets/images/phase/NN.png';
import EE from '@/assets/images/phase/EE.png';

// 记录未匹配的项
const unmatchArray = ref([]);

const props = defineProps({
  // 图片的宽度
  imageWidth: String,
  //图片的高度
  imageHeight: String,

  duration : Number,
  // 当前的值
  value: [String],
  // 当未找到匹配的数据时，显示value
  showValue: {
    type: Boolean,
    default: false,
  },
  separator: {
    type: String,
    default: "_",
  }
});

const values = computed(() => {
  if (props.value === null || typeof props.value === 'undefined' || props.value === '') return [];
  return Array.isArray(props.value) ? props.value.map(item => '' + item) : String(props.value).split(props.separator);
});


</script>

<style scoped>
.el-tag + .el-tag {
  margin-left: 10px;
}
</style>
