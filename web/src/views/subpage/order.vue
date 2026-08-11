<template>
<div class='order-page'>
  <van-nav-bar
    :title="lang('商城订单')"
    left-arrow
    :border="false"
    fixed
    @click-left="handleBack"
  />
  <div class="page-main">
    <van-tabs v-model:active="active" @change="changeTab">
      <van-tab :title="lang('全部')" name="0">
        <div class="order-list" v-if="list.length > 0">
          <div class="order-item" v-for="item in list">
            <div class="order-header">
              <span>{{ item.createdAt }}</span>
              <span>{{ orderTypeOption[item.orderType] }}</span>
            </div>
            <div class="order-product">
              <div class="order-product-image">
                <img :src="`/images/${item.six}`" />
              </div>
              <div class="order-product-info">
                <p>{{ item.four || '-' }}</p>
                <p>{{ item.five || '-' }}</p>
              </div>
            </div>
            <div class="order-address">
              <p>{{lang('收件人')}}：{{ item.name || '-' }}  {{ item.phone}}</p>
              <p>{{lang('收货地址')}}：{{ (item.country + item.province + item.city + item.area + item.detail) || '-' }}</p>
            </div>
            <div class="order-footer">
              {{lang('订单金额')}}：{{ item.amount }}
            </div>
          </div>
        </div>
        <van-empty :description="lang('暂无数据')" v-else :image="emptyImage" />
      </van-tab>
      <van-tab :title="lang('待发货')" name="1">
        <div class="order-list" v-if="list.length > 0">
          <div class="order-item" v-for="item in list">
            <div class="order-header">
              <span>{{ item.createdAt }}</span>
              <span>{{ orderTypeOption[item.orderType] }}</span>
            </div>
            <div class="order-product">
              <div class="order-product-image">
                <img :src="`/images/${item.six}`" />
              </div>
              <div class="order-product-info">
                <p>{{ item.four || '-' }}</p>
                <p>{{ item.five || '-' }}</p>
              </div>
            </div>
            <div class="order-address">
              <p>{{lang('收件人')}}：{{ item.name || '-' }}  {{ item.phone}}</p>
              <p>{{lang('收货地址')}}：{{ (item.country + item.province + item.city + item.area + item.detail) || '-' }}</p>
            </div>
            <div class="order-footer">
              {{lang('订单金额')}}：{{ item.amount }}
            </div>
          </div>
        </div>
        <van-empty :description="lang('暂无数据')" v-else :image="emptyImage" />
      </van-tab>
      <van-tab :title="lang('待收货')" name="2">
        <div class="order-list" v-if="list.length > 0">
          <div class="order-item" v-for="item in list">
            <div class="order-header">
              <span>{{ item.createdAt }}</span>
              <span>{{ orderTypeOption[item.orderType] }}</span>
            </div>
            <div class="order-product">
              <div class="order-product-image">
                <img :src="`/images/${item.six}`" />
              </div>
              <div class="order-product-info">
                <p>{{ item.four || '-' }}</p>
                <p>{{ item.five || '-' }}</p>
              </div>
            </div>
            <div class="order-address">
              <p>{{lang('收件人')}}：{{ item.name || '-' }}  {{ item.phone}}</p>
              <p>{{lang('收货地址')}}：{{ (item.country + item.province + item.city + item.area + item.detail) || '-' }}</p>
            </div>
            <div class="order-footer">
              {{lang('订单金额')}}：{{ item.amount }}
            </div>
          </div>
        </div>
        <van-empty :description="lang('暂无数据')" v-else :image="emptyImage" />
      </van-tab>
      <van-tab :title="lang('已完成')" name="3">
        <div class="order-list" v-if="list.length > 0">
          <div class="order-item" v-for="item in list">
            <div class="order-header">
              <span>{{ item.createdAt }}</span>
              <span>{{ orderTypeOption[item.orderType] }}</span>
            </div>
            <div class="order-product">
              <div class="order-product-image">
                <img :src="`/images/${item.six}`" />
              </div>
              <div class="order-product-info">
                <p>{{ item.four || '-' }}</p>
                <p>{{ item.five || '-' }}</p>
              </div>
            </div>
            <div class="order-address">
              <p>{{lang('收件人')}}：{{ item.name || '-' }}  {{ item.phone}}</p>
              <p>{{lang('收货地址')}}：{{ (item.country + item.province + item.city + item.area + item.detail) || '-' }}</p>
            </div>
            <div class="order-footer">
              {{lang('订单金额')}}：{{ item.amount }}
            </div>
          </div>
        </div>
        <van-empty :description="lang('暂无数据')" v-else :image="emptyImage" />
      </van-tab>
    </van-tabs>
    <Pagination
      v-if="allPage > 0"
      v-model="page"
      :page-count="allPage"
      mode="simple"
      @change="handlePage"
    />
  </div>
</div>
</template>
<script setup>
import userPerson from "@/pinia/person";
import request from "@/tools/request";
import { useRouter, useRoute } from 'vue-router'
import { Pagination } from "vant"
import emptyImage from '../../assets/images/custom-empty-image.png'
import lang from '@/i18n/index'

const route = useRoute()

const router = useRouter()
const person = userPerson();
const active = $ref('0')
const allPage = $ref(0)
const page = $ref(1)
const list = $ref([])

console.log(route.params.id)

const orderTypeOption = {
  "1": lang('未发货'),
  "2": lang('已发货'),
  "3": lang('已送达')
}

const getOrder = async () => {
  let fetchUrl = null

  if (route.params.id === '1') {
    fetchUrl = "app_server/order_three_list"
  } else {
    fetchUrl = "app_server/order_two_list"
  }

  await request.get(fetchUrl, {
    params: {
      page,
      orderType: active
    }
  }).then((res) => {
    if (res.status === 'ok') {
      allPage = Math.ceil(res.count / 10);
      list = res.list
    } else {
      showFailToast(res.status)
    }
  })
}

getOrder()

const changeTab = (value) => {
  list = []
  page = 1
  allPage = 0
  getOrder()
}

const handlePage = (value) => {
  page = value
  getOrder()
}

const handleBack = () => {
  router.back()
}

</script>
<style lang='less' scoped>
  .order-page {
    min-height: 100vh;
    .page-main {
      padding-top: 50px;
    }
    /deep/ .van-tab__panel {
      padding: 20px 0;
    }
    .order-list {
      display: flex;
      flex-direction: column;
      gap: 10px;
      padding: 0 15px;
      .order-item {
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: #0D1B2A;
        border-radius: 6px;
        padding: 15px;
        .order-header {
          border-bottom: 1px solid #222;
          padding-bottom: 10px;
          display: flex;
          justify-content: space-between;
          color: #999;
        }
        .order-product {
          display: flex;
          align-items: flex-start;
          gap: 20px;
          border-bottom: 1px solid #222;
          padding-bottom: 10px;
          .order-product-image {
            width: 80px;
            height: 80px;
            background: #444;
            flex-shrink: 0;
            border-radius: 6px;
            overflow: hidden;
            img {
              width: 80px;
              height: 80px;
            }
          }
          .order-product-info {
            display: flex;
            flex-direction: column;
            gap: 15px;
            color: #CCC;
          }
        }
        .order-address {
          border-bottom: 1px solid #222;
          padding-bottom: 10px;
          color: #CCC;
          display: flex;
          flex-direction: column;
          gap: 10px;
        }
      }
      .order-footer {
        text-align: right;
        font-size: 14px;
        color: #CCC;
      }
    }
  }
</style>