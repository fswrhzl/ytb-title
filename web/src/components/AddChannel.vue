<template>
    <div class="modal-overlay" @click="closeModal">
        <div class="modal-container" @click.stop>
            <div class="modal-content">
                <div class="modal-header">
                    <h2 class="modal-title">📺 {{ props.modalType === "add" ? "新增" : "编辑" }}频道</h2>
                    <button
                        @click="closeModal"
                        class="close-btn"
                        :class="{ 'close-btn-hover': isCloseHovered }"
                        @mouseenter="isCloseHovered = true"
                        @mouseleave="isCloseHovered = false"
                    >
                        ✕
                    </button>
                </div>

                <div class="modal-body">
                    <div class="input-group">
                        <label class="input-label">🏷️ 频道名称：</label>
                        <input v-model="channelName" type="text" placeholder="📺 请输入频道名称..." class="channel-input" @keyup.enter="confirmAdd" @keyup.esc="closeModal" />
                    </div>

                    <div class="input-group">
                        <label class="input-label">🎯 选择标签（多选）：</label>
                        <div v-if="availableTags.length > 0" class="tags-container">
                            <div v-for="tag in availableTags" :key="tag.name" class="tag-item">
                                <label class="tag-label">
                                    <input type="checkbox" :value="tag.id" v-model="selectedTags" class="tag-checkbox" />
                                    <span :class="['tag-span', selectedTags.some((item) => item === tag.id) ? 'tag-selected' : 'tag-unselected']">
                                        {{ tag.name }}
                                    </span>
                                </label>
                            </div>
                        </div>
                        <div v-else class="no-tags-message">🏷️ 暂无可用标签，请先添加标签</div>
                    </div>

                    <div class="button-group">
                        <button
                            @click="confirmAdd"
                            :disabled="!channelName.trim()"
                            class="confirm-btn"
                            :class="{
                                'confirm-btn-disabled': !channelName.trim(),
                            }"
                        >
                            ✅ 确认
                        </button>
                        <button @click="closeModal" class="cancel-btn">❌ 取消</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";

const props = defineProps(["modalType", "editedChannel"]);
const channelName = ref(props.modalType === "add" ? "" : props.editedChannel.name);
const isCloseHovered = ref(false);
const selectedTags = ref(props.modalType === "add" ? [] : (props.editedChannel.tags ?? []));
const availableTags = ref(localStorage.getItem("tags") ? JSON.parse(localStorage.getItem("tags")) : []);
const emit = defineEmits(["close", "flushChannels"]);
const closeModal = () => {
    emit("close");
};
const confirmAdd = () => {
    if (!channelName.value.trim()) {
        return;
    }
    let apiInfo = {
        url: "",
        method: "",
        error: "",
    };
    let requestData = {
        name: channelName.value.trim(),
        tags: selectedTags.value,
    };
    if (props.modalType === "add") {
        apiInfo.url = "/api/channels";
        apiInfo.method = "post";
        apiInfo.error = "新增频道失败，请稍后重试。";
    } else {
        apiInfo.url = `/api/channels/${props.editedChannel.id}`;
        apiInfo.method = "put";
        apiInfo.error = "编辑频道失败，请稍后重试。";
        requestData.id = props.editedChannel.id;
    }
    // 新增或编辑频道
    axios
        .request({
            method: apiInfo.method,
            url: apiInfo.url,
            data: requestData,
        })
        .then((response) => {
            alert(response.data.message);
            if (response.data.status !== "success") {
                return;
            }
            selectedTags.value = []; // 清空选中标签
            channelName.value = ""; // 清空频道名称输入框
            closeModal();
            emit("flushChannels");
        })
        .catch((error) => {
            console.log(apiInfo.error, error);
            alert(apiInfo.error);
        });
};

// 生命周期
onMounted(() => {
    // 自动聚焦到输入框
    setTimeout(() => {
        const input = document.querySelector(".channel-input");
        if (input) input.focus();
    }, 100);
});
</script>

<style scoped>
@import url("../assets/add-channel.css");
</style>
