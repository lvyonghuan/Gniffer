<template>
  <div style="width: 100vw; height: 100vh; margin: 0; padding: 0; overflow: hidden;">
    <Dashboard v-if="list && list.length > 0" :list="list" style="width: 100%; height: 100%;" />
    <div v-else>
      Loading... 
    </div>
  </div>
</template>

<style>
html, body {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>

<script>
import Dashboard from './dashboard.vue'
import axios from 'axios'

export default {
  name: 'App',
  components: {
    Dashboard
  },
  data() {
    return {
      list: []
    }
  },
  methods: {
    async getList() {
      try {
        const response = await axios.get('http://127.0.0.1:8080/list')
        if (response.data && response.data.cards && Array.isArray(response.data.cards)) {
          this.list = response.data.cards
          // console.log('List:', this.list)
          // console.log('List length:', this.list.length)
        } else {
          console.error('Response data is not in the expected format:', response.data)
        }
      } catch (error) {
        console.error('Error fetching list:', error)
      }
    }
  },
  mounted() {
    this.getList()
  }
}
</script>