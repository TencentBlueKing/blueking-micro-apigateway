<template>
  <div class="collapse">
    <div class="header">
      <div class="prefix">
        <bk-checkbox
          v-model="enabled"
          :disabled="disabled"
          @change="handleCheckboxChanged"
        />
      </div>
      <div
        class="title"
        @click="handleTitleClicked"
      >
        <div class="name">
          {{ name }}
        </div>
        <div class="desc">
          <bk-overflow-title
            class="overflow-hidden"
            type="tips"
          >
            {{ desc }}
          </bk-overflow-title>
        </div>
      </div>
      <div
        :class="{ 'collapsed': !collapsed }"
        class="suffix"
      >
        <icon
          color="#979BA5"
          name="down-shape"
          size="10"
        />
      </div>
    </div>
    <div
      v-show="!collapsed"
      class="content"
    >
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';

interface IProps {
  name?: string
  desc?: string
  disabled?: boolean
}

const enabled = defineModel<boolean>({ default: false });

const collapsed = defineModel<boolean>('collapsed', { default: true });

const {
  name = '',
  desc = '',
  disabled = false,
} = defineProps<IProps>();

const handleCheckboxChanged = (checked: boolean) => {
  collapsed.value = !checked;
};

const handleTitleClicked = () => {
  if (disabled) {
    collapsed.value = true;
    return;
  }
  collapsed.value = !collapsed.value;
};

</script>

<style lang="scss" scoped>
.collapse {
  width: 100%;

  .header {
    display: flex;
    align-items: center;
    height: 40px;
    padding-left: 8px;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    background: #fafbfd;

    .prefix {
      display: flex;
      align-items: center;
      justify-content: center;
      padding-right: 8px;
    }

    .title {
      display: flex;
      align-items: center;
      width: calc(100% - 50px);
      cursor: pointer;

      .name {
        font-size: 14px;
        font-weight: 700;
        line-height: 20px;
        flex-shrink: 0;
        margin-right: 12px;
        color: #4d4f56;
      }

      .desc {
        font-size: 12px;
        line-height: 20px;
        width: calc(100% - 68px);
        color: #4d4f56;
      }
    }

    .suffix {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: auto;
      padding: 8px;
      transition: transform;

      &.collapsed {
        transform: rotate(180deg);
      }
    }
  }

  .content {
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 2px;
    background-color: #ffffff;
    padding-block: 16px;
  }
}

</style>
