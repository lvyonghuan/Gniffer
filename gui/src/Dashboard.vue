<template>
  <el-container>
    <!-- 侧边栏 -->
    <el-aside width="200px" class="aside">
      <el-select v-model="selectedOption" placeholder="请选择网卡">
        <el-option 
          v-for="item in optionList" 
          :key="item.name"
          :label="item.description"
          :value="item.name">
        </el-option>
      </el-select>

      <el-container>
        <el-row class="button-group" justify="center" style="margin-top: 20px">
          <el-button 
            type="success" 
            class="custom-button" 
            style="margin: 5px" 
            :disabled="isCapturing" 
            @click="startCapture">开始捕获</el-button>
          <el-button 
            type="warning" 
            class="custom-button" 
            style="margin: 5px" 
            :disabled="!isCapturing" 
            @click="restartCapture">重新捕获</el-button>
          <el-button 
            type="danger" 
            class="custom-button" 
            style="margin: 5px" 
            :disabled="!isCapturing" 
            @click="stopCapture">停止捕获</el-button>
        </el-row>
      </el-container>
    </el-aside>

    <!-- 主体内容 -->
    <el-container>
      <!-- 顶部过滤器 -->
      <el-header style="background-color: #f5f7fa">
        <el-row :gutter="20" style="margin-top: 10px">
          <el-col :span="18">
            <el-input placeholder="请输入过滤信息......" v-model="searchText"></el-input>
          </el-col>
          <el-col :span="6">
            <el-button type="primary" @click="applyFilter">启用</el-button>
          </el-col>
        </el-row>
      </el-header>

      <!-- 信息接收区 -->
      <el-main>
        <Sniffer 
          :selectedOption="selectedOption"
          :searchText="searchText"
          :isCapturing="isCapturing"
        />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import axios from 'axios';
import Sniffer from './Sniffer.vue';

export default {
  components: {
    Sniffer
  },
  data() {
    return {
      selectedOption: '',
      searchText: '',
      optionList: [],
      isCapturing: false
    }
  },
  methods: {
    startCapture() {
      this.isCapturing = true;
    },
    restartCapture() {
      this.isCapturing = true;
    },
    stopCapture() {
      this.isCapturing = false;
    },
    applyFilter() {
      const filter = this.searchText;
      axios.get(`http://127.0.0.1:8080/setFilter`, {
        params: {
          filter: filter
        }
      })
      .then(response => {
        console.log('Filter applied:', response.data);
      })
      .catch(error => {
        console.error('Error applying filter:', error);
      });
    }
  },
  props: {
    list: {
      type: Array,
      required: true
    }
  },
  created() {
    this.optionList = this.list.map(item => {
      return {
        name: item.name,
        description: item.description
      }
    })
  }
}
</script>

<style scoped>
.aside {
  background-color: #9b9191;
  padding: 10px;
  height: 100vh;
  box-shadow: 2px 0 12px 0 rgba(0, 0, 0, 0.1);
}

.custom-button {
  width: 100px;
  color: white;
}
</style>