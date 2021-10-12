<template>
  <div class="columns">
    <!-- navbar -->
    <div class="column has-background-grey-dark mx-2" style="max-width: 10rem">
      <router-link to="/report" class="button is-info mb-2">
        <span class="icon">
          <i class="fas fa-analytics"></i>
        </span>
        <span>Reports</span>
      </router-link>
      <router-link to="/monitors" class="button is-info mb-2">
        <span class="icon">
          <i class="fas fa-tools"></i>
        </span>
        <span>Monitors</span>
      </router-link>
    </div>

    <!-- main view-->
    <div class="column">
      <!-- <router-view v-slot="{ Component }">
        <keep-alive>
          <component :is="Component" />
        </keep-alive>
      </router-view> -->
      <router-view />
    </div>
  </div>
</template>

<style scoped>
.columns {
  height: 100%;
  margin-top: 0;
  margin-bottom: 0;
}
.column {
  height: 100%;
}
.button {
  width: 100%;
}
</style>
