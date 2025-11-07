<template>
    <div class="modal-overlay" @click="closeModal">
        <div class="modal-container" @click.stop>
            <!-- Header区域 -->
            <header class="modal-header">
                <h2 class="modal-title">🏷️ 管理标签</h2>
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

            <!-- Content区域 -->
            <main class="modal-main">
                <div class="content-area">
                    <div v-if="tags.length === 0" class="no-tags-message">
                        <p class="info-message">📭 暂无标签，请先添加标签</p>
                    </div>
                    <div v-else class="tags-list">
                        <div v-for="(tag, index) in tags" :key="tag.id" class="tag-item">
                            <div class="tag-info">
                                <span class="tag-name">{{ tag.name }}</span>
                                <span class="tag-channels">关联频道: {{ getChannelNames(tag.channels).join(", ") || "无" }}</span>
                            </div>
                            <button
                                @click="deleteTag(tag.id, index)"
                                class="delete-btn"
                                :class="{ 'delete-btn-hover': hoveredTagId === tag.id }"
                                @mouseenter="hoveredTagId = tag.id"
                                @mouseleave="hoveredTagId = null"
                                title="删除标签"
                            >
                                🗑️
                            </button>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import axios from "axios";

const emit = defineEmits(["close", "flushTags"]);
const tags = ref(localStorage.getItem("tags") ? JSON.parse(localStorage.getItem("tags")) : []);
const isCloseHovered = ref(false);
const hoveredTagId = ref(null);
const channels = ref(localStorage.getItem("channels") ? JSON.parse(localStorage.getItem("channels")) : []);
// 方法
const closeModal = () => {
    emit("close");
};
// 获取标签关联的频道名称
const getChannelNames = (channelIds) => {
    if (!channelIds || !channels.value?.length) return [];
    return channels.value
        .filter((channel) => {
            return channelIds.includes(channel.id);
        })
        .map((channel) => channel.name);
};

// 删除标签
const deleteTag = async (tagId, index) => {
    if (!confirm("确定要删除这个标签吗？")) {
        return;
    }
    try {
        const response = await axios.delete(`/api/tags/${tagId}`);
        alert(response.data.message);
        if (response.data.status !== "success") {
            return;
        }
        tags.value.splice(index, 1);
        emit("flushTags");
    } catch (error) {
        console.error("删除标签失败:", error);
        alert("删除标签失败，请稍后重试。");
    }
};
</script>

<style scoped>
@import url("../assets/manage-tag.css");
</style>
