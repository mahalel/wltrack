// Import dependencies
import { Chart } from 'chart.js/auto';
import 'htmx.org';

// Make Chart available globally
window.Chart = Chart;

// Set up all application functionality after DOM is loaded
document.addEventListener("DOMContentLoaded", function () {
  // Select elements based on page
  setupWorkoutForm();
  setupCharts();
  setupWeightInputs();
  setupHtmxForms();

  // Set today as default date for new workout form
  const dateInput = document.getElementById("date");
  if (dateInput) {
    const today = new Date();
    const yyyy = today.getFullYear();
    let mm = today.getMonth() + 1;
    let dd = today.getDate();

    if (dd < 10) dd = "0" + dd;
    if (mm < 10) mm = "0" + mm;

    dateInput.value = `${yyyy}-${mm}-${dd}`;
  }

  // Set up HTMX event listeners
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (evt.detail.xhr.getResponseHeader("HX-Redirect")) {
      window.location.href = evt.detail.xhr.getResponseHeader("HX-Redirect");
      evt.preventDefault();
    }

    // Handle HX-Refresh for full page refreshes
    if (evt.detail.xhr.getResponseHeader("HX-Refresh")) {
      setTimeout(function () {
        window.location.reload();
      }, 800);
      evt.preventDefault(); // Prevent the swap since we're refreshing the page
    }
  });

  document.body.addEventListener("htmx:afterRequest", function (evt) {
    // Backup handler for HX-Refresh
    if (evt.detail.xhr.getResponseHeader("HX-Refresh")) {
      setTimeout(function () {
        window.location.reload();
      }, 800);
    }
  });
});

// Handle dynamic exercise and set range addition in workout form
function setupWorkoutForm() {
  const exerciseList = document.getElementById("exercise-list");
  const addExerciseBtn = document.getElementById("add-exercise");

  // No workout form on this page
  if (!exerciseList || !addExerciseBtn) {
    return;
  }

  // Add exercise to workout
  addExerciseBtn.addEventListener("click", function () {
    const exerciseTemplate = document.getElementById("exercise-template");
    const exerciseId = document.querySelectorAll(".exercise-item").length + 1;
    let exerciseContent = exerciseTemplate.innerHTML;
    
    // Replace placeholder IDs with unique IDs
    exerciseContent = exerciseContent.replace(/EXERCISE_INDEX/g, exerciseId);
    
    // Create a wrapper div and set innerHTML
    const exerciseDiv = document.createElement("div");
    exerciseDiv.classList.add("exercise-item", "mb-6", "border", "border-gray-200", "rounded-lg", "p-4");
    exerciseDiv.innerHTML = exerciseContent;
    
    // Add the new exercise to the list
    exerciseList.appendChild(exerciseDiv);
    
    // Attach event listeners to new elements
    attachSetRangeListeners(exerciseDiv);
    attachExerciseRemoveListeners(exerciseDiv);
    attachSetRangeInputHandlers(exerciseDiv);
    
    // Set focus on the new exercise input
    const exerciseInput = exerciseDiv.querySelector("input[type='text']");
    if (exerciseInput) {
      exerciseInput.focus();
    }
  });

  // Initial setup of existing exercises
  attachSetRangeListeners(document);
  attachExerciseRemoveListeners(document);
  attachSetRangeInputHandlers(document);
  updateAllSetRanges();
}

function attachSetRangeListeners(container) {
  // Add a new set to an exercise
  container.querySelectorAll(".add-set").forEach(button => {
    // Remove existing event listeners to prevent duplication
    const newButton = button.cloneNode(true);
    button.parentNode.replaceChild(newButton, button);
    
    newButton.addEventListener("click", function () {
      const exerciseDiv = this.closest(".exercise-item");
      const setList = exerciseDiv.querySelector(".set-list");
      const exerciseIndex = this.dataset.exerciseIndex || 
                          exerciseDiv.querySelector("[data-exercise-index]").dataset.exerciseIndex;
      
      const setTemplate = document.getElementById("set-template");
      let setContent = setTemplate.innerHTML;
      
      // Get the current number of sets for proper indexing
      const setIndex = setList.querySelectorAll(".set-item").length + 1;
      
      // Replace placeholders with actual indices
      setContent = setContent.replace(/EXERCISE_INDEX/g, exerciseIndex);
      setContent = setContent.replace(/SET_INDEX/g, setIndex);
      
      // Create a wrapper div
      const setDiv = document.createElement("div");
      setDiv.classList.add("set-item", "mb-2", "flex", "items-center", "space-x-3");
      setDiv.innerHTML = setContent;
      
      // Add the new set to the list
      setList.appendChild(setDiv);
      
      // Attach event handlers to new input fields
      attachSetRangeInputHandlers(setDiv);
      
      // Set focus on the new weight input
      const weightInput = setDiv.querySelector("input[name^='weight']");
      if (weightInput) {
        weightInput.focus();
      }
      
      // Update the range counter for this exercise
      updateSetRangeCounters(exerciseDiv);
    });
  });
}

function attachExerciseRemoveListeners(container) {
  container.querySelectorAll(".remove-exercise").forEach(button => {
    // Remove existing event listeners
    const newButton = button.cloneNode(true);
    button.parentNode.replaceChild(newButton, button);
    
    newButton.addEventListener("click", function () {
      const exerciseItem = this.closest(".exercise-item");
      exerciseItem.remove();
      
      // Update all form elements to ensure proper indexing
      updateAllSetRanges();
    });
  });
}

function attachSetRangeInputHandlers(container) {
  // Add handlers for all relevant input types
  const inputSelectors = [
    "input[name^='weight']", 
    "input[name^='reps']", 
    "select[name^='rpe']", 
    "input[type='range']"
  ];
  
  inputSelectors.forEach(selector => {
    container.querySelectorAll(selector).forEach(input => {
      // Remove any existing event listeners
      const newInput = input.cloneNode(true);
      input.parentNode.replaceChild(newInput, input);
      
      // For range sliders, update the counter and link to actual input
      if (newInput.type === "range") {
        const counter = newInput.previousElementSibling;
        const actualInput = container.querySelector(`input[name='${newInput.dataset.target}']`) ||
                          container.querySelector(`select[name='${newInput.dataset.target}']`);
        
        // Update counter and actual input when slider changes
        newInput.addEventListener("input", function () {
          if (counter) {
            counter.textContent = this.value;
          }
          if (actualInput) {
            actualInput.value = this.value;
          }
        });
        
        // Set initial counter value
        if (counter && newInput.value) {
          counter.textContent = newInput.value;
        }
        
        // Link actual input back to the slider
        if (actualInput) {
          actualInput.addEventListener("input", function () {
            newInput.value = this.value;
            if (counter) {
              counter.textContent = this.value;
            }
          });
          
          // Set initial slider value from actual input
          if (actualInput.value) {
            newInput.value = actualInput.value;
            if (counter) {
              counter.textContent = actualInput.value;
            }
          }
        }
      }
      
      // For weight and reps inputs, update related range sliders
      if (newInput.name && (newInput.name.includes('weight') || newInput.name.includes('reps') || newInput.name.includes('rpe'))) {
        newInput.addEventListener("input", function () {
          const exerciseItem = this.closest(".exercise-item");
          const rangeSliderId = this.name.replace(/[\[\]]/g, '-') + '-range';
          const rangeSlider = exerciseItem.querySelector(`input[data-target='${this.name}']`);
          const counter = rangeSlider ? rangeSlider.previousElementSibling : null;
          
          if (rangeSlider) {
            rangeSlider.value = this.value;
            if (counter) {
              counter.textContent = this.value;
            }
          }
        });
      }
    });
  });
}

function updateSetRangeCounters(exerciseDiv) {
  if (!exerciseDiv) return;
  
  const setList = exerciseDiv.querySelector(".set-list");
  const setCounter = exerciseDiv.querySelector(".set-counter");
  
  if (setList && setCounter) {
    const setCount = setList.querySelectorAll(".set-item").length;
    setCounter.textContent = `Sets: ${setCount}`;
  }
}

function updateAllSetRanges() {
  // Get all exercises
  const exercises = document.querySelectorAll(".exercise-item");
  
  // For each exercise, update indices
  exercises.forEach((exercise, exerciseIndex) => {
    // Update exercise index
    const newExerciseIndex = exerciseIndex + 1;
    
    // Update data attributes
    exercise.querySelectorAll("[data-exercise-index]").forEach(el => {
      el.dataset.exerciseIndex = newExerciseIndex;
    });
    
    // Update input names
    exercise.querySelectorAll("input[name], select[name]").forEach(input => {
      if (input.name.includes('exercises[')) {
        input.name = input.name.replace(/exercises\[\d+\]/, `exercises[${newExerciseIndex}]`);
      }
    });
    
    // Update set indices within this exercise
    const sets = exercise.querySelectorAll(".set-item");
    sets.forEach((set, setIndex) => {
      const newSetIndex = setIndex + 1;
      
      // Update input names with new indices
      set.querySelectorAll("input[name], select[name]").forEach(input => {
        if (input.name.includes('sets[')) {
          input.name = input.name.replace(/sets\[\d+\]/, `sets[${newSetIndex}]`);
        }
      });
    });
    
    // Update the set counter
    updateSetRangeCounters(exercise);
  });
}

function setupCharts() {
  // Call page-specific chart setup functions
  setupHomePageChart();
  setupExerciseDetailChart();
  
  // Any shared chart configuration can go here
}

function setupHtmxForms() {
  // Add flash message fade-out
  document.querySelectorAll(".flash-message").forEach(message => {
    setTimeout(() => {
      message.style.opacity = "0";
      setTimeout(() => {
        message.style.display = "none";
      }, 500); // Allow time for fade out animation
    }, 5000); // Display for 5 seconds
  });
  
  // Add confirmation for delete actions
  document.querySelectorAll(".confirm-delete").forEach(element => {
    element.addEventListener("click", function(e) {
      if (!confirm("Are you sure you want to delete this item? This cannot be undone.")) {
        e.preventDefault();
        return false;
      }
    });
  });
  
  // Handle form submission animations
  document.querySelectorAll("form").forEach(form => {
    form.addEventListener("submit", function() {
      const submitButton = this.querySelector("button[type='submit']");
      const loadingIndicator = this.querySelector(".loading-indicator");
      
      if (submitButton && loadingIndicator) {
        submitButton.disabled = true;
        submitButton.classList.add("opacity-50");
        loadingIndicator.classList.remove("hidden");
      }
    });
  });
  
  // Reset form state after HTMX request completes
  document.body.addEventListener("htmx:afterRequest", function(evt) {
    const form = evt.detail.elt.closest("form");
    
    if (form) {
      const submitButton = form.querySelector("button[type='submit']");
      const loadingIndicator = form.querySelector(".loading-indicator");
      
      if (submitButton && loadingIndicator) {
        submitButton.disabled = false;
        submitButton.classList.remove("opacity-50");
        loadingIndicator.classList.add("hidden");
      }
    }
  });
}

function setupHomePageChart() {
  const chartCanvas = document.getElementById("progress-chart");
  
  if (!chartCanvas) {
    return; // Not on a page with the progress chart
  }
  
  // Get workout data from data attribute
  let workoutData = [];
  try {
    const dataElement = document.getElementById("chart-data");
    if (dataElement && dataElement.dataset.workouts) {
      workoutData = JSON.parse(dataElement.dataset.workouts);
    }
  } catch (e) {
    console.error("Error parsing workout data:", e);
    return;
  }
  
  // Format data for Chart.js
  const dates = [];
  const weights = {};
  const exerciseList = [];
  
  // First pass: Collect all exercise names
  workoutData.forEach(workout => {
    workout.exercises.forEach(exercise => {
      if (!exerciseList.includes(exercise.name)) {
        exerciseList.push(exercise.name);
      }
    });
  });
  
  // Sort exercises alphabetically
  exerciseList.sort();
  
  // Initialize weight datasets
  exerciseList.forEach(exercise => {
    weights[exercise] = {};
  });
  
  // Second pass: Collect dates and weights
  workoutData.forEach(workout => {
    const date = new Date(workout.date);
    const formattedDate = `${date.getMonth() + 1}/${date.getDate()}`;
    
    if (!dates.includes(formattedDate)) {
      dates.push(formattedDate);
    }
    
    workout.exercises.forEach(exercise => {
      // Find the heaviest set for this exercise
      let maxWeight = 0;
      
      exercise.sets.forEach(set => {
        if (set.weight > maxWeight) {
          maxWeight = set.weight;
        }
      });
      
      // Store the max weight for this exercise on this date
      if (maxWeight > 0) {
        weights[exercise.name][formattedDate] = maxWeight;
      }
    });
  });
  
  // Sort dates chronologically
  dates.sort((a, b) => {
    const [monthA, dayA] = a.split("/").map(n => parseInt(n));
    const [monthB, dayB] = b.split("/").map(n => parseInt(n));
    
    if (monthA !== monthB) {
      return monthA - monthB;
    }
    return dayA - dayB;
  });
  
  function updateChart() {
    // Get selected exercises from checkboxes
    const selectedExercises = [];
    document.querySelectorAll(".exercise-checkbox:checked").forEach(cb => {
      selectedExercises.push(cb.value);
    });
    
    // Generate datasets
    const datasets = selectedExercises.map((exercise, index) => {
      const color = getColorForIndex(index);
      
      const data = dates.map(date => {
        return weights[exercise][date] || null;
      });
      
      return {
        label: exercise,
        data: data,
        borderColor: color,
        backgroundColor: color + '33', // Add transparency for fill
        fill: false,
        tension: 0.1,
        pointRadius: 4,
        pointHoverRadius: 6
      };
    });
    
    // Create or update chart
    if (window.progressChart) {
      window.progressChart.data.labels = dates;
      window.progressChart.data.datasets = datasets;
      window.progressChart.update();
    } else {
      window.progressChart = new Chart(chartCanvas, {
        type: 'line',
        data: {
          labels: dates,
          datasets: datasets
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            title: {
              display: true,
              text: 'Progress Over Time',
              font: {
                size: 16
              }
            },
            legend: {
              position: 'top',
            },
            tooltip: {
              callbacks: {
                title: function(tooltipItems) {
                  return 'Date: ' + tooltipItems[0].label;
                },
                label: function(context) {
                  return context.dataset.label + ': ' + context.raw + ' lbs';
                }
              }
            }
          },
          scales: {
            y: {
              title: {
                display: true,
                text: 'Weight (lbs)'
              },
              beginAtZero: false
            },
            x: {
              title: {
                display: true,
                text: 'Date'
              }
            }
          }
        }
      });
    }
  }
  
  // Create exercise selector
  const exerciseSelector = document.getElementById("exercise-selector");
  
  if (exerciseSelector) {
    // Create checkboxes for each exercise
    exerciseList.forEach(exercise => {
      const checkboxId = 'exercise-' + exercise.replace(/\s+/g, '-').toLowerCase();
      
      const checkboxDiv = document.createElement('div');
      checkboxDiv.className = 'flex items-center space-x-2';
      
      const checkbox = document.createElement('input');
      checkbox.type = 'checkbox';
      checkbox.id = checkboxId;
      checkbox.className = 'exercise-checkbox';
      checkbox.value = exercise;
      
      // Check first 5 exercises by default
      if (exerciseList.indexOf(exercise) < 5) {
        checkbox.checked = true;
      }
      
      const label = document.createElement('label');
      label.htmlFor = checkboxId;
      label.className = 'text-sm';
      label.textContent = exercise;
      
      // When checkbox changes, update chart
      checkbox.addEventListener('change', updateChart);
      
      checkboxDiv.appendChild(checkbox);
      checkboxDiv.appendChild(label);
      exerciseSelector.appendChild(checkboxDiv);
    });
  }
  
  // Helper function to get a color based on index
  function getColorForIndex(index) {
    const colors = [
      '#3B82F6', // Blue
      '#EF4444', // Red
      '#10B981', // Green
      '#F97316', // Orange
      '#8B5CF6', // Purple
      '#EC4899', // Pink
      '#14B8A6', // Teal
      '#F59E0B', // Amber
      '#6366F1', // Indigo
      '#84CC16'  // Lime
    ];
    
    return colors[index % colors.length];
  }
  
  // Initial chart update
  updateChart();
}

function setupWeightInputs() {
  document.querySelectorAll("input[type='number'][name*='weight']").forEach(input => {
    // Add event listener to handle decimal values properly
    input.addEventListener('blur', function() {
      let value = parseFloat(this.value);
      
      if (!isNaN(value)) {
        // Round to one decimal place
        value = Math.round(value * 10) / 10;
        this.value = value;
      }
    });
    
    // Add keydown handler for increment/decrement keys
    input.addEventListener('keydown', function(e) {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
        e.preventDefault();
        
        let value = parseFloat(this.value) || 0;
        let step = e.shiftKey ? 10 : (e.ctrlKey || e.metaKey) ? 5 : 2.5;
        
        if (e.key === 'ArrowDown') {
          step = -step;
        }
        
        value += step;
        value = Math.round(value * 10) / 10;
        
        // Ensure positive weight
        if (value < 0) value = 0;
        
        this.value = value;
        
        // Trigger input event for sliders and other elements
        this.dispatchEvent(new Event('input', { bubbles: true }));
      }
    });
  });
}

function setupExerciseDetailChart() {
  const chartCanvas = document.getElementById("exercise-progress-chart");
  
  if (!chartCanvas) {
    return; // Not on the exercise detail page
  }
  
  // Get workout data from data attribute
  let progressData = [];
  try {
    const dataElement = document.getElementById("exercise-chart-data");
    if (dataElement && dataElement.dataset.progress) {
      progressData = JSON.parse(dataElement.dataset.progress);
    }
  } catch (e) {
    console.error("Error parsing exercise progress data:", e);
    return;
  }
  
  // Format data for Chart.js
  const dates = [];
  const maxWeights = [];
  const volumeData = [];
  const setData = [];
  
  // Process the data
  progressData.forEach(entry => {
    const date = new Date(entry.date);
    const formattedDate = `${date.getMonth() + 1}/${date.getDate()}`;
    
    dates.push(formattedDate);
    maxWeights.push(entry.max_weight);
    volumeData.push(entry.volume);
    setData.push(entry.sets);
  });
  
  // Create the chart
  const chart = new Chart(chartCanvas, {
    type: 'bar',
    data: {
      labels: dates,
      datasets: [
        {
          type: 'line',
          label: 'Max Weight (lbs)',
          data: maxWeights,
          borderColor: '#3B82F6', // Blue
          backgroundColor: 'rgba(59, 130, 246, 0.2)',
          borderWidth: 2,
          tension: 0.1,
          pointRadius: 4,
          pointBackgroundColor: '#3B82F6',
          yAxisID: 'y',
        },
        {
          type: 'bar',
          label: 'Volume (lbs)',
          data: volumeData,
          backgroundColor: 'rgba(16, 185, 129, 0.5)', // Green with transparency
          borderColor: 'rgba(16, 185, 129, 1)',
          borderWidth: 1,
          yAxisID: 'y1',
        },
        {
          type: 'line',
          label: 'Sets',
          data: setData,
          borderColor: '#F97316', // Orange
          backgroundColor: 'rgba(249, 115, 22, 0.2)',
          borderWidth: 2,
          pointRadius: 4,
          pointBackgroundColor: '#F97316',
          borderDash: [5, 5],
          tension: 0.1,
          yAxisID: 'y2',
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: {
          display: true,
          text: 'Exercise Progress Over Time',
          font: {
            size: 16
          }
        },
        legend: {
          position: 'top',
        },
        tooltip: {
          callbacks: {
            title: function(tooltipItems) {
              return 'Date: ' + tooltipItems[0].label;
            },
            label: function(context) {
              let label = context.dataset.label || '';
              let value = context.raw;
              
              if (label.includes('Weight') || label.includes('Volume')) {
                return label + ': ' + value + ' lbs';
              } else {
                return label + ': ' + value;
              }
            }
          }
        }
      },
      scales: {
        x: {
          title: {
            display: true,
            text: 'Date'
          }
        },
        y: {
          position: 'left',
          title: {
            display: true,
            text: 'Weight (lbs)'
          },
          beginAtZero: false,
        },
        y1: {
          position: 'right',
          title: {
            display: true,
            text: 'Volume (lbs)'
          },
          beginAtZero: true,
          grid: {
            drawOnChartArea: false,
          }
        },
        y2: {
          position: 'right',
          title: {
            display: true,
            text: 'Sets'
          },
          beginAtZero: true,
          grid: {
            drawOnChartArea: false,
          },
          ticks: {
            stepSize: 1
          }
        }
      }
    }
  });
}