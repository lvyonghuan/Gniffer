<template>
    <el-table
        ref="table"
        :data="tableData"
        style="width: 100%; height: 400px"
        virtual-scrolling
        :height="400"
        :item-size="40"
        @scroll="handleScroll"
        @row-click="handleRowClick"
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
        <div v-if="selectedRow" style="width: 100%; height: 200px; overflow: auto;">
            <el-collapse>
                <el-collapse-item v-for="(item, index) in selectedRow.children" :key="index" :title="item.name">
                    <div style="display: flex; justify-content: space-between;">
                        <pre>{{ item.info }}</pre>
                        <pre style="margin-right: 60px;">{{ item.hex.match(/.{1,2}/g).join(' ') }}</pre>
                    </div>
                </el-collapse-item>
            </el-collapse>
        </div>
    </el-container>
</template>

<script>
export default {
    data() {
        return {
            tableData: [],
            maxIndex: 0,
            currentStartIndex: 0,
            currentEndIndex: 25,
            selectedRow: null,
        }
    },
    name: 'Sniffer',
    props: {
        selectedOption: {
            type: String,
            required: true
        },
        searchText: {
            type: String,
            required: true
        },
        isCapturing: {
            type: Boolean,
            required: true
        }
    },
    watch: {
        selectedOption(newVal, oldVal) {

        },
        searchText(newVal, oldVal) {
            console.log('searchText changed:', oldVal, '->', newVal);
        },
        async isCapturing(newVal, oldVal) {
            if (newVal) {
                await fetch(`http://127.0.0.1:8080/listen?netCard=${this.selectedOption}`);

                this.tableData = [];
                this.maxIndex = 0;
                this.currentStartIndex = 0;
                this.currentEndIndex = 25;
                this.startCapture();
            } else {
                this.stopCapture();
            }
        }
    },

    methods: {
        startCapture() {
            this.pollTimer = setInterval(async () => {
                const start = Math.max(0, this.maxIndex - 25);
                const end = Math.max(25, this.maxIndex);
                const response = await fetch(`http://127.0.0.1:8080/require?start=${start}&end=-1`);
                const { data, index } = await response.json();
                this.tableData = data.reverse();
                this.maxIndex = index;
                this.currentStartIndex = start;
                this.currentEndIndex = end;
            }, 1000);
        },
        async handleScroll({ scrollTop }) {
            if (scrollTop === 0 && this.isCapturing) {
                if (!this.pollTimer) {
                    this.startCapture();
                }
            } else {
                if (this.pollTimer) {
                    clearInterval(this.pollTimer);
                    this.pollTimer = null;
                }
                if (scrollTop + this.$refs.table.$el.clientHeight >= this.$refs.table.$el.scrollHeight) {
                    const lastItem = this.tableData[this.tableData.length - 1];
                    const end = lastItem.id - 1;
                    const start = lastItem.id - 50;
                    const response = await fetch(`http://127.0.0.1:8080/require?start=${start}&end=${end}`);
                    const { data } = await response.json();
                    this.tableData = this.tableData.concat(data.reverse());
                }
            }
        },
        async stopCapture() {
            if (this.pollTimer) {
                clearInterval(this.pollTimer);
                this.pollTimer = null;
            }
            await fetch(`http://127.0.0.1:8080/stop?netCard=${this.selectedOption}`);
        },
        handleRowClick(row) {
            this.selectedRow = row;
        }
    },
}
</script>