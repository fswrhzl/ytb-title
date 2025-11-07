<script setup>
import CreateTitle from "./components/CreateTitle.vue";
import { onUnmounted } from "vue";

onUnmounted(() => {
    localStorage.removeItem("channels");
    localStorage.removeItem("tags");
});
</script>
<template>
    <CreateTitle />
</template>

<style scoped></style>
