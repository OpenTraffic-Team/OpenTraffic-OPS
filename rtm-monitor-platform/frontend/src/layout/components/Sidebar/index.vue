<template>
  <div class="sidebar"  :style="{ backgroundColor: sideTheme === 'theme-dark' ? variables.menuBackground : variables.menuLightBackground }">
    <logo v-if="showLogo" :collapse="false"/>
    <el-scrollbar class="scrollbar" wrap-class="scrollbar-wrapper">
      <el-menu
        ref="menuRef"
        :default-active="activeMenu"
        :collapse="false"
        :background-color="sideTheme === 'theme-dark' ? variables.menuBackground : variables.menuLightBackground"
        :text-color="sideTheme === 'theme-dark' ? variables.menuColor : variables.menuLightColor"
        :unique-opened="false"
        :collapse-transition="false"
        mode="vertical"
      >
        <sidebar-item
          v-for="(route, index) in sidebarRouters"
          :key="route.path + index"
          :item="route"
          :base-path="route.path"
        />
      </el-menu>
    </el-scrollbar>
    <div class="right-menu">
      <div class="avatar-container">
        <el-dropdown @command="handleCommand" class="right-menu-item hover-effect" trigger="click">
          <div class="avatar-wrapper">
            <img :src="userStore.avatar" class="user-avatar" />
            <el-icon><caret-bottom /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <router-link to="/user/profile">
                <el-dropdown-item>个人中心</el-dropdown-item>
              </router-link>
              <el-dropdown-item divided command="logout">
                <span>退出登录</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
     </div>
  </div>
</template>

<script setup>
import Logo from './Logo'
import SidebarItem from './SidebarItem'
import variables from '@/assets/styles/variables.module.scss'
import useAppStore from '@/store/modules/app'
import useSettingsStore from '@/store/modules/settings'
import usePermissionStore from '@/store/modules/permission'
import useUserStore from "@/store/modules/user.js";
import {CaretBottom} from "@element-plus/icons-vue";
import {ElMessageBox} from "element-plus";
import {useRoute} from "vue-router";

const route = useRoute();
const appStore = useAppStore()
const userStore = useUserStore()
const settingsStore = useSettingsStore()
const permissionStore = usePermissionStore()

const sidebarRouters =  computed(() => permissionStore.sidebarRouters);
const showLogo = computed(() => settingsStore.sidebarLogo);
const sideTheme = computed(() => settingsStore.sideTheme);
const theme = computed(() => settingsStore.theme);

const menuRef = ref(null);

function expandAllMenus() {
  if (!menuRef.value) return;
  const subMenus = menuRef.value.subMenus;
  if (subMenus) {
    Object.keys(subMenus).forEach(index => {
      menuRef.value.open(index);
    });
  }
}

watch(sidebarRouters, () => {
  nextTick(() => expandAllMenus());
}, { deep: true });

onMounted(() => {
  nextTick(() => expandAllMenus());
});

const activeMenu = computed(() => {
  const { meta, path } = route;
  // if set path, the sidebar will highlight the path you set
  if (meta.activeMenu) {
    return meta.activeMenu;
  }
  return path;
})
function handleCommand(command) {
  switch (command) {
    case "logout":
      logout();
      break;
    default:
      break;
  }
}

function logout() {
  ElMessageBox.confirm('确定注销并退出系统吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    userStore.logOut().then(() => {
      location.href = '/index';
    })
  }).catch(() => { });
}
</script>
<style lang='scss' scoped>
.sidebar{
  display:flex;
  flex-direction: column;
  height: 100%;

  .scrollbar{
    flex: 1;
    width: 100%;

    /*去除导航菜单下划白线*/
    .el-menu.el-menu--vertical{
      border: none;
    }
  }

  .right-menu {
    width: 100%;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-top: 1px solid rgba(255, 255, 255, 0.1);

    &:focus {
      outline: none;
    }

    .avatar-container {
      margin: 0;

      .avatar-wrapper {
        position: relative;
        display: flex;
        align-items: center;
        gap: 8px;

        .user-avatar {
          cursor: pointer;
          width: 40px;
          height: 40px;
          border-radius: 10px;
        }

        i {
          cursor: pointer;
          font-size: 12px;
        }
      }
    }
  }
}
</style>
