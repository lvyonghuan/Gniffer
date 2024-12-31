<template>
    <el-table
        :data="displayedData"
        style="width: 100%"
        v-loading="loading"
        element-loading-text="加载中..."
        element-loading-spinner="el-icon-loading"
    >
        <el-table-column prop="id" label="序号" width="80" />
        <el-table-column prop="time" label="时间" width="180" />
        <el-table-column prop="source" label="始发地" width="180" />
        <el-table-column prop="destination" label="目的地" width="180" />
        <el-table-column prop="protocol" label="协议" width="120" />
        <el-table-column prop="length" label="长度" width="100" />
    </el-table>

    <!-- 具体信息展示区域 -->
    <el-container>
        <!-- 可以根据需要扩展 -->
    </el-container>
</template>

<script>
export default {
    data() {
        return {
            tableData: [], // 全部捕获数据
            buffer: [], // 数据缓冲区
            loading: false, // 加载状态
            updateInterval: null, // 定时器引用
            maxRows: 10000, // 最大显示行数
        };
    },
    name: "Sniffer",
    props: {
        selectedOption: {
            type: String,
            required: true,
        },
        searchText: {
            type: String,
            required: true,
        },
        isCapturing: {
            type: Boolean,
            required: true,
        },
    },
    computed: {
        displayedData() {
            // 根据搜索文本动态过滤数据
            return this.searchText
                ? this.tableData.filter((item) =>
                      Object.values(item)
                          .join(" ")
                          .toLowerCase()
                          .includes(this.searchText.toLowerCase())
                  )
                : this.tableData;
        },
    },
    watch: {
        selectedOption(newVal, oldVal) {
            console.log("selectedOption changed:", oldVal, "->", newVal);
        },
        searchText(newVal, oldVal) {
            console.log("searchText changed:", oldVal, "->", newVal);
        },
        isCapturing(newVal, oldVal) {
            if (newVal) {
                this.startCapture();
            } else {
                this.stopCapture();
            }
        },
    },
    methods: {
        startCapture() {
            if (this.ws) {
                this.ws.close();
            }

            this.loading = true;
            const ws = new WebSocket(
                `ws://127.0.0.1:8080/listen?netCard=${this.selectedOption}`
            );

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.buffer.push(data);
            };

            ws.onerror = (err) => {
                console.error("WebSocket error:", err);
                this.loading = false;
            };

            ws.onclose = () => {
                this.loading = false;
            };

            this.ws = ws;

            // 设置定时器批量刷新表格数据
            this.updateInterval = setInterval(() => {
                if (this.buffer.length > 0) {
                    const newData = this.buffer.splice(0, this.buffer.length);
                    // 合并数据并限制长度
                    this.tableData = [
                        ...newData,
                        ...this.tableData,
                    ]
                        .slice(0, this.maxRows)
                        .sort((a, b) => a.id - b.id); // 按序号排序
                }
            }, 100);
        },
        stopCapture() {
            console.log("stopCapture");

            if (this.ws) {
                this.ws.close();
                this.ws = null;
            }

            if (this.updateInterval) {
                clearInterval(this.updateInterval);
                this.updateInterval = null;
            }

            this.loading = false;
        },
    },
};
</script>
