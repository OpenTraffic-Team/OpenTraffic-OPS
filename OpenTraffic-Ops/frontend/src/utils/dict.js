import { ref, toRefs } from 'vue'

/**
 * 内置常用字典数据（简化版，不再依赖后端字典接口）
 */
const builtInDicts = {
  sys_normal_disable: [
    { label: '正常', value: '0', elTagType: 'success', elTagClass: '' },
    { label: '停用', value: '1', elTagType: 'danger', elTagClass: '' }
  ],
  sys_user_sex: [
    { label: '男', value: '0', elTagType: '', elTagClass: '' },
    { label: '女', value: '1', elTagType: 'warning', elTagClass: '' },
    { label: '未知', value: '2', elTagType: 'info', elTagClass: '' }
  ],
  sys_show_hide: [
    { label: '显示', value: '0', elTagType: 'primary', elTagClass: '' },
    { label: '隐藏', value: '1', elTagType: 'warning', elTagClass: '' }
  ],
  sys_yes_no: [
    { label: '是', value: 'Y', elTagType: 'primary', elTagClass: '' },
    { label: '否', value: 'N', elTagType: 'danger', elTagClass: '' }
  ],
  sys_notice_type: [
    { label: '通知', value: '1', elTagType: 'warning', elTagClass: '' },
    { label: '公告', value: '2', elTagType: 'success', elTagClass: '' }
  ],
  sys_notice_status: [
    { label: '正常', value: '0', elTagType: 'primary', elTagClass: '' },
    { label: '关闭', value: '1', elTagType: 'danger', elTagClass: '' }
  ],
  sys_oper_type: [
    { label: '其它', value: '0', elTagType: '', elTagClass: '' },
    { label: '新增', value: '1', elTagType: 'primary', elTagClass: '' },
    { label: '修改', value: '2', elTagType: 'warning', elTagClass: '' },
    { label: '删除', value: '3', elTagType: 'danger', elTagClass: '' },
    { label: '授权', value: '4', elTagType: 'success', elTagClass: '' },
    { label: '导出', value: '5', elTagType: 'info', elTagClass: '' },
    { label: '导入', value: '6', elTagType: 'info', elTagClass: '' },
    { label: '强退', value: '7', elTagType: 'danger', elTagClass: '' },
    { label: '生成代码', value: '8', elTagType: 'success', elTagClass: '' },
    { label: '清空数据', value: '9', elTagType: 'danger', elTagClass: '' }
  ],
  sys_common_status: [
    { label: '成功', value: '0', elTagType: 'success', elTagClass: '' },
    { label: '失败', value: '1', elTagType: 'danger', elTagClass: '' }
  ]
}

/**
 * 获取字典数据（内置硬编码，不再请求后端）
 */
export function useDict(...args) {
  const res = ref({})
  return (() => {
    args.forEach((dictType) => {
      res.value[dictType] = builtInDicts[dictType] || []
    })
    return toRefs(res.value)
  })()
}
