<template>
    <div class="app-container">
        <!-- 主要内容区域 -->
        <div class="main-content">
            <div class="content-wrapper">
                <!-- 标题 -->
                <div class="hero-section">
                    <h2 class="hero-title">油管标题生成器</h2>
                    <div class="hero-menu">
                        <button @click="toggleHeroMenu" class="hero-menu-btn" :class="{ 'hero-menu-btn-active': isHeroMenuOpen }">
                            {{ isHeroMenuOpen ? "✕" : "☰" }}
                        </button>
                        <div v-show="isHeroMenuOpen" class="hero-menu-dropdown">
                            <ul class="hero-menu-list">
                                <li>
                                    <button @click="showAddChannel" class="hero-menu-link">📺 新增频道</button>
                                </li>
                                <li>
                                    <button @click="showAddTag" class="hero-menu-link">🏷️ 新增标签</button>
                                </li>
                                <li>
                                    <button @click="showEditTag" class="hero-menu-link">🏷️ 管理标签</button>
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
                <p class="hero-subtitle">&nbsp;</p>
                <!-- 输入表单 -->
                <div class="form-container">
                    <div class="form-content">
                        <!-- 视频主题输入 -->
                        <div class="input-group">
                            <label class="input-label">🎯 视频标题：</label>
                            <input v-model="videoTheme" type="text" placeholder="🎬 输入你的视频标题" class="theme-input" />
                        </div>

                        <!-- 频道选择 -->
                        <div class="input-group">
                            <label class="input-label">📺 选择频道（单选）：</label>
                            <div class="tags-container">
                                <div v-for="channel in availableChannels.channels" :key="channel.id" class="tag-item" @contextmenu.prevent="showContextMenu($event, channel)">
                                    <label class="tag-label">
                                        <input type="radio" :value="channel.id" v-model="selectedChannel" class="tag-checkbox" />
                                        <span :class="['tag-span', selectedChannel === channel.id ? 'tag-selected' : 'tag-unselected']">
                                            {{ channel.name }}
                                        </span>
                                    </label>
                                    <button @click.stop="deleteChannel(channel.id)" class="delete-btn" title="删除频道">✕</button>
                                </div>
                                <div v-if="availableChannels.channels.length === 0">
                                    <div v-if="!availableChannels.status">
                                        <p class="info-message">⚠️ 获取频道失败，请稍后重试。</p>
                                    </div>
                                    <div v-else>
                                        <p class="info-message">📭 暂无频道</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 生成按钮 -->
                        <div class="button-container">
                            <button
                                @click="generateTitle"
                                :disabled="!videoTheme.trim() || selectedChannel === ''"
                                class="generate-btn"
                                :class="{
                                    'generate-btn-disabled': !videoTheme.trim() || selectedChannel === '',
                                }"
                            >
                                🚀 生成标题
                            </button>
                        </div>

                        <!-- 生成的标题显示 -->
                        <div v-if="generatedTitle" class="result-container">
                            <h3 class="result-title">🎉 生成的标题：</h3>
                            <div class="titles-list">
                                <div class="title-item" @click="copyTitle(generatedTitle)">
                                    <span class="title-text">{{ generatedTitle }}</span>
                                    <span class="copy-hint">📋 点击复制</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <!-- AddChannel 模态框 -->
        <AddChannel v-if="isAddChannelVisible" @close="hideAddChannel" @flushChannels="getChannels" :modalType="channelModalType" :editedChannel="editedChannel" />
        <!-- AddTag 模态框 -->
        <AddTag v-if="isAddTagVisible" @close="hideAddTag" @flushTags="getTags" />
        <!-- ManageTag 模态框 -->
        <ManageTag v-if="isManageTagVisible" @close="hideEditTag" @flushTags="getTags" />
        <!-- 右键菜单 -->
        <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }" @click="hideContextMenu">
            <div class="context-menu-item" @click="editChannel(contextMenu.channel)">✏️ 编辑</div>
        </div>
        <!-- 右键菜单专用遮罩层，用于关闭右键菜单 -->
        <div v-if="contextMenu.visible" class="context-menu-overlay" @click="hideContextMenu"></div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";
import AddChannel from "./AddChannel.vue";
import AddTag from "./AddTag.vue";
import ManageTag from "./ManageTag.vue";

const isHeroMenuOpen = ref(false);
const isAddChannelVisible = ref(false);
const isAddTagVisible = ref(false);
const isManageTagVisible = ref(false);
const videoTheme = ref("");
const selectedChannel = ref("");
const generatedTitle = ref("");
const availableTags = ref([]);
// 右键菜单状态
const contextMenu = ref({
    visible: false,
    x: 0,
    y: 0,
    channel: null,
});
// 可选频道
const availableChannels = ref({
    channels: [],
    status: false, // 数据状态，false为数据不可用
});
const channelModalType = ref("add"); // 新增频道或编辑频道
const editedChannel = ref(null); // 编辑的频道对象
axios.defaults.headers = {
    "Content-Type": "application/json",
};
// 方法
const toggleHeroMenu = () => {
    isHeroMenuOpen.value = !isHeroMenuOpen.value;
};
const showAddChannel = () => {
    isAddChannelVisible.value = true;
    isHeroMenuOpen.value = false; // 关闭菜单
};
const hideAddChannel = () => {
    isAddChannelVisible.value = false;
    channelModalType.value = "add";
    editedChannel.value = null;
};

const showAddTag = () => {
    isAddTagVisible.value = true;
    isHeroMenuOpen.value = false; // 关闭菜单
};

const hideAddTag = () => {
    isAddTagVisible.value = false;
};

const showEditTag = () => {
    isManageTagVisible.value = true;
    isHeroMenuOpen.value = false; // 关闭菜单
};

const hideEditTag = () => {
    isManageTagVisible.value = false;
};

const generateTitle = async () => {
    if (!videoTheme.value.trim() || selectedChannel.value === "") {
        return;
    }
    await axios
        .post("/api/generate-title", {
            theme: videoTheme.value.trim(),
            channel: selectedChannel.value,
        })
        .then((response) => {
            alert(response.data.message);
            if (response.data.status !== "success") {
                return;
            }
            generatedTitle.value = response.data.title;
        })
        .catch((error) => {
            console.error("生成标题失败:", error);
            alert("生成标题失败，请稍后重试。");
        });
};

const copyTitle = async (title) => {
    try {
        await navigator.clipboard.writeText(title);
        alert("标题已复制到剪贴板！");
    } catch (err) {
        // 如果浏览器不支持 clipboard API，使用传统方法
        const textArea = document.createElement("textarea");
        textArea.value = title;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand("copy");
        document.body.removeChild(textArea);
        alert("标题已复制到剪贴板！");
    }
};
const getTags = async () => {
    await axios
        .get("/api/tags")
        .then((response) => {
            if (response.data.status !== "success") {
                alert(response.data.message);
                return;
            }
            availableTags.value = response.data.tags ?? [];
            availableTags.value.status = true;
            localStorage.setItem("tags", JSON.stringify(availableTags.value));
        })
        .catch((error) => {
            console.error("获取标签失败:", error);
            availableTags.value.status = false;
        });
};
const getChannels = () => {
    axios
        .get("/api/channels")
        .then((response) => {
            if (response.data.status !== "success") {
                alert(response.data.message);
                return;
            }
            availableChannels.value.channels = response.data.channels ?? [];
            availableChannels.value.status = true;
            localStorage.setItem("channels", JSON.stringify(availableChannels.value.channels));
        })
        .catch((error) => {
            console.error("获取频道失败:", error);
            availableChannels.value.status = false;
        });
};
// 删除频道
const deleteChannel = async (channelId) => {
    if (confirm("确定要删除这个频道吗？")) {
        try {
            // 重新获取频道列表
            availableChannels.value.channels = availableChannels.value.channels.filter((channel) => channel.id !== channelId);
            localStorage.setItem("channels", JSON.stringify(availableChannels.value.channels));
            // 如果删除的是当前选中的频道，清空选择
            if (selectedChannel.value === channelId) {
                selectedChannel.value = "";
            }
            const response = await axios.delete(`/api/channels/${channelId}`);
            if (response.data.status !== "success") {
                alert(response.data.message);
                return;
            }
            alert("频道删除成功！");
        } catch (error) {
            console.error("删除频道失败:", error);
            alert("删除频道失败，请稍后重试。");
        }
    }
};

// 显示右键菜单
const showContextMenu = (event, channel) => {
    contextMenu.value = {
        visible: true,
        x: event.clientX,
        y: event.clientY,
        channel: channel,
    };
};

// 隐藏右键菜单
const hideContextMenu = () => {
    contextMenu.value.visible = false;
};

// 编辑频道
const editChannel = (channel) => {
    hideContextMenu();
    isAddChannelVisible.value = true;
    channelModalType.value = "edit"; // 设置为编辑模式
    editedChannel.value = channel; // 存储编辑的频道对象
};
onMounted(() => {
    // 初始化时获取所有频道信息
    getTags();
    getChannels();
});
</script>

<style scoped>
@import url("../assets/create-title.css");
</style>
