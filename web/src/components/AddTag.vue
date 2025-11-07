<template>
    <div class="modal-overlay" @click="closeModal">
        <div class="modal-container" @click.stop>
            <!-- Header区域 -->
            <header class="modal-header">
                <h2 class="modal-title">🏷️ 新增标签</h2>
                <button
                    @click="closeModal"
                    class="close-btn"
                    :class="{ 'close-btn-hover': isCloseHovered }"
                    @mouseenter="isCloseHovered = true"
                    @mouseleave="isCloseHovered = false"
                >
                    ✕
                </button>
            </header>
            <!-- Main区域 -->
            <main class="modal-main">
                <div class="form-content">
                    <div class="input-group">
                        <label for="tagInput" class="input-label">📝 标签名称：</label>
                        <input ref="tagInput" v-model="tagName" type="text" id="tagInput" class="tag-input" placeholder="🏷️ 输入标签名称..." @keyup.enter="confirmAdd" />
                    </div>

                    <div class="input-group" style="max-height: 200px; overflow-y: auto">
                        <label class="input-label">📺 所属频道（多选）：</label>
                        <div class="radio-group">
                            <label v-for="channel in tagChannels" :key="channel.id" class="radio-label">
                                <input type="checkbox" :value="channel.id" v-model="selectedChannels" class="radio-input" />
                                <span :class="['radio-span', selectedChannels.includes(channel.id) ? 'radio-selected' : 'radio-unselected']">
                                    {{ channel.name }}
                                </span>
                            </label>
                        </div>
                    </div>
                </div>
            </main>

            <!-- Footer区域 -->
            <footer class="modal-footer">
                <button
                    @click="confirmAdd"
                    :disabled="!tagName.trim() || !selectedChannels.length"
                    class="confirm-btn"
                    :class="{
                        'confirm-btn-disabled': !tagName.trim() || !selectedChannels.length,
                    }"
                >
                    ✅ 确定添加
                </button>
            </footer>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import axios from "axios";

// 定义事件
const emit = defineEmits(["close", "flushTags"]);
// 响应式数据
const tagName = ref("");
const selectedChannels = ref([]);
const isCloseHovered = ref(false);
const tagInput = ref(null);
// 标签所属频道
const tagChannels = ref(localStorage.getItem("channels") ? JSON.parse(localStorage.getItem("channels")) : []);
// 方法
const closeModal = () => {
    emit("close");
};
const confirmAdd = () => {
    if (!tagName.value.trim()) {
        alert("请输入标签名称！");
        return;
    }
    if (!selectedChannels.value.length) {
        alert("请选择所属频道！");
        return;
    }
    axios
        .post("/api/tags", {
            name: tagName.value.trim(),
            channels: selectedChannels.value,
        })
        .then((response) => {
            if (response.data.status !== "success") {
                alert(response.data.message);
                return;
            }
            alert("标签添加成功！");
            // closeModal();
            emit("flushTags");
        })
        .catch((error) => {
            console.error("添加标签失败:", error);
            alert("添加标签失败，请稍后重试。");
        });
};

// 键盘事件处理
const handleKeydown = (e) => {
    if (e.key === "Escape") {
        closeModal();
    }
};

// 生命周期
onMounted(() => {
    document.addEventListener("keydown", handleKeydown);
    // 自动聚焦到输入框
    if (tagInput.value) {
        tagInput.value.focus();
    }
});

onUnmounted(() => {
    document.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
@import url("../assets/add-tag.css");
</style>
