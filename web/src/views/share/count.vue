<template>
  <f7-page name="count">
    <ChildrenHeader />
    <div class="share-main">
      <f7-toolbar class="tab-menu" bottom tabbar>
        <f7-link tab-link="#tab-a" tab-link-active>{{ $t('count.myNodes') }}</f7-link>
        <f7-link tab-link="#tab-b">{{ $t('count.myIncome') }}</f7-link>
        <f7-link tab-link="#tab-c">{{ $t('count.matrix') }}</f7-link>
      </f7-toolbar>
      <f7-tabs animated>
        <f7-tab id="tab-a" class="tab-content" tab-active>
          <f7-block>
            <div class="node-head">
              <div class="node-head-item">
                <p>{{ $t('count.level') }}</p>
                <p>{{ levelType[userinfo.level] }}</p>
              </div>
              <div class="node-head-item">
                <p>{{ $t('count.sharedNodes') }}</p>
                <p>{{ userinfo.locationNum }}</p>
              </div>
              <div class="node-head-item">
                <p>{{ $t('count.totalPerformance') }}</p>
                <p>{{userinfo.total}}</p>
              </div>
              <div class="node-head-item">
                <p>{{ $t('count.regionalPerformance') }}</p>
                <p>{{userinfo.max}}</p>
              </div>
              <div class="node-head-item">
                <p>{{ $t('count.smallAreaPerformance') }}</p>
                <p>{{userinfo.min}}</p>
              </div>
              <div class="node-head-item">
                <p>{{ $t('count.referrer') }}</p>
                <p style="font-size: 13px">{{ formatAddress(userinfo.inviteUserAddress) }}</p>
              </div>
            </div>
            <!-- <div class="node-list">
              <div class="null-content" v-if="userinfo.listRecommend.length === 0">{{ $t('common.noData') }}</div>
              <ul v-else>
                <li v-for="(item, index) in userinfo.listRecommend" :key="index">{{ item.address }}</li>
              </ul>
            </div> -->
          </f7-block>
        </f7-tab>
        <f7-tab id="tab-b" class="tab-content">
          <f7-block>
            <div class="content-box">
              <div class="income-box">
                <div class="income-main">
                  <p>{{ $t('count.pendingIncome') }}</p>
                  <p>${{userinfo.amountGetSub || 0}}</p>
                </div>
                <div class="income-footer">
                  <div class="income-footer-item">
                    <p>{{ $t('count.node') }}</p>
                    <p>{{userinfo.buy || 0}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.pendingOutput') }}</p>
                    <p>{{userinfo.amountGetSub || 0}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.produced') }}</p>
                    <p>{{userinfo.amountGet || 0}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.exitCount') }}</p>
                    <p>{{userinfo.outNum || 0}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.staticIncome') }}</p>
                    <p>{{userinfo.location}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.directIncome') }}</p>
                    <p>{{userinfo.recommend}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.directAcceleration') }}</p>
                    <p>{{userinfo.recommendTwo}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.teamIncome') }}</p>
                    <p>{{userinfo.team}}</p>
                  </div>
                  <div class="income-footer-item">
                    <p>{{ $t('count.networkIncome') }}</p>
                    <p>{{userinfo.all}}</p>
                  </div>
                </div>
              </div>
            </div>
            <div class="income-list">
              <div class="income-list-header">
                <ul class="header-list">
                  <li :class="{'active': activeTab === item[0]}" v-for="(item, index) in menuType" :key="index" @click="changeTab(item[0])">{{ item[1] }}</li>
                </ul>
              </div>
              <div class="income-list-main">
                <div class="income-list-item" v-for="(item, index) in rewardList" :key="index">
                  <div class="income-list-item-info">
                    <p>
                      <span v-if="!['8'].includes(activeTab)">{{ $t('community.usdtAmount') }}: {{ item.amount }}</span>
                      <span v-if="!['1'].includes(activeTab)">{{ $t('community.brc20Amount') }}: {{ item.amountTwo }}</span>
                    </p>
                    <p style="font-size: 13px;">
                      <span v-if="!['1', '2', '7', '8'].includes(activeTab)">{{ formatAddress(item.address) }}</span>
                      <span v-if="!['1', '2', '3', '7', '8'].includes(activeTab)">{{ $t('community.generation') }}: {{ item.num }}</span>
                    </p>
                    <p>{{ item.createdAt }}</p>
                  </div>
                  <div class="income-list-item-money">{{ item.reward }}</div>
                </div>
                <div class="empty-data" v-if="rewardList.length === 0"></div>
                <Pagination
                  v-model="page"
                  :page-count="allPageCount"
                  mode="simple"
                  @change="getRewardList"
                />
              </div>
            </div>
          </f7-block>
        </f7-tab>
        <f7-tab id="tab-c" class="tab-content">
          <f7-block>
            <a-tree
              v-if="treeData.length > 0"
              v-model:expandedKeys="expandedKeys"
              v-model:selectedKeys="selectedKeys"
              :load-data="onLoadData"
              :tree-data="treeData"
            />
          </f7-block>
        </f7-tab>
      </f7-tabs>
    </div>
  </f7-page>
</template>
<script setup lang="ts">
import ChildrenHeader from '../../components/header/childrenHeader.vue'
import userPerson from "@/pinia/person";
import type { TreeProps } from 'ant-design-vue';
import { Pagination } from "vant";
import { f7, f7ready } from 'framework7-vue'
import { computed, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n'
import request from "@/tools/request";
const person = userPerson();
const { t: $t } = useI18n()
const userinfo = $computed(() => person.userinfo);
const address = $computed(() => person.address);

let page = $ref(1);
let allPageCount = $ref(3);
let activeTab: string = $ref('1');
let pickerDevice: any = $ref(null);
let rewardList: any[] = $ref([]);

const menuType = computed(() => [
  ['1', $t('count.subscribe')],
  ['2', $t('count.staticIncome')],
  ['3', $t('count.directIncome')],
  ['4', $t('count.directAccelerationIncome')],
  ['5', $t('count.generationTeamIncome')],
  ['7', $t('count.networkIncome')],
  ['8', `${$t('count.gift')} BRC20`]
])

const getRewardList = async (page: number = 1) => {
  const res: any = await request.get("app_server/reward_list", {
    params: {
      page,
      reqType: activeTab
    }
  });

  allPageCount = Math.ceil(res.count / 10);
  rewardList = res.list
}

getRewardList()

const changeTab = (type: string) => {
  activeTab = type;

  page = 1
  getRewardList()
}

onMounted(() => {
})

onBeforeUnmount(() => {
})

const levelType = computed<Record<string, string>>(() => ({
  '-1': $t('count.inactive'),
  '0': $t('count.node'),
  '1': 'v1',
  '2': 'v2',
  '3': 'v3',
  '4': 'v4',
  '5': 'v5'
}))

const expandedKeys = $ref<string[]>([]);
const selectedKeys = $ref<string[]>([]);
let treeData: any = $ref([]);

const onLoadData: TreeProps['loadData'] = (treeNode: any) => {
  return new Promise<void>(async (resolve) => {
    if (treeNode.dataRef.children) {
      resolve();
      return;
    }

    const res: any = await request.get(`app_server/recommend_list?address=${treeNode.dataRef.address}`);

    // res.area = [
    //   {
    //     address: formatAddress("aaaa12378123hjdfowis88883"),
    //     locationId: "3",
    //     countLow: 2,
    //   },
    //   {
    //     address: formatAddress("32423998f8uijkjkrejkw2223"),
    //     locationId: "4",
    //     countLow: 0
    //   }
    // ]
    setTimeout(() => {
      treeNode.dataRef.children = res.recommends.map((item: any, index: number) => {
        const hasChildren = item.hasChildren != null ? !!item.hasChildren : true
        const tag = item.activated === false ? $t('count.inactive') : `${$t('common.quantity')}: ${item.amount}`
        return {
          title: `${formatAddress(item.address)}---(${tag})`,
          key: `${treeNode.eventKey}-${index}`,
          amount: item.amount,
          address: item.address,
          isLeaf: !hasChildren
        }
      })
      treeData = [...treeData];
      resolve();
    }, 1000);
  });
};

const formatAddress = (value: string) => {
  const frontSix = value.slice(0, 6);
  const backSix = value.slice(-4);
  const middle = '...';
  return frontSix + middle + backSix;
}

const getUserArea = async () => {
  const res: any = await request.get(`app_server/recommend_list?address=${address}`);
  // res.area = [
  //   {
  //     address: formatAddress("aaaa12378123hjdfowis88883"),
  //     locationId: "3",
  //     countLow: 2,
  //   },
  //   {
  //     address: formatAddress("32423998f8uijkjkrejkw2223"),
  //     locationId: "4",
  //     countLow: 0
  //   }
  // ]
  treeData = res.recommends.map((item: any, index: number) => {
    const hasChildren = item.hasChildren != null ? !!item.hasChildren : Number(item.countLow || 0) > 0
    const tag = item.activated === false ? $t('count.inactive') : `${$t('common.quantity')}: ${item.amount}`
    return {
      title: `${formatAddress(item.address)}---(${tag})`,
      key: index,
      address: item.address,
      isLeaf: !hasChildren,
    }
  })
}

getUserArea()


</script>
<style scoped lang="less">
@import "./styles/index.less";
</style>
