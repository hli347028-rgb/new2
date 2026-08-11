<template>
    <div class="login-view">
        <div class="loading">
            <div class="loading-spinner"></div>
            <div class="loading-text">
                <div v-if="person.loadAccount">{{lang('common.contractVerifying')}}</div>
                <div v-else>{{lang('common.walletConnecting')}}</div>
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
import userPerson from "@/pinia/person";
import lang from '@/i18n/index'
const person = userPerson();
</script>
<style scoped lang="less">
.login-view {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: #000;
}

.loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
}

.loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: #1597E5;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.loading-text {
    color: rgba(255, 255, 255, 0.8);
    font-size: 14px;
}
</style>