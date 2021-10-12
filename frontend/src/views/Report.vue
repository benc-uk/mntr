<template>
  <div>
    <canvas id="chart" ref="chartCanvas"></canvas>
  </div>
</template>

<script lang="ts">
import { onMounted, defineComponent, ref, Ref } from 'vue'
import * as chartjs from 'chart.js'
let chart

export default defineComponent({
  setup() {
    // Weird way of initalizing the ref, due to TypeScript
    // This is the DOM canvas object with `ref="chartCanvas"`
    const chartCanvas = ref({}) as Ref<HTMLCanvasElement>

    onMounted(async () => {
      chart = new chartjs.Chart(chartCanvas.value, {
        // The type of chart we want to create
        type: 'line',

        // The data for our dataset
        data: {
          // labels: [
          //   'January',
          //   'February',
          //   'March',
          //   'April',
          //   'May',
          //   'June',
          //   'July'
          // ],
          // datasets: [
          //   {
          //     label: 'My First dataset',
          //     backgroundColor: 'rgb(255, 99, 132)',
          //     borderColor: 'rgb(255, 99, 132)',
          //     data: [0, 10, 5, 2, 20, 30, 45]
          //   }
          // ]
        },

        // Configuration options go here
        options: {}
      })
    })

    // expose to template
    return {
      chartCanvas
    }
  }
})
</script>

<style lang="css" scoped></style>
