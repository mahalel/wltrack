// Import Chart.js
import { Chart } from 'chart.js/auto';

// Make Chart available globally
window.Chart = Chart;

// Import our existing JavaScript functionality after DOM is loaded
document.addEventListener("DOMContentLoaded", function() {
  // Load our existing functionality from the original main.js
  const script = document.createElement('script');
  script.src = '/static/js/original-main.js';
  document.head.appendChild(script);
});