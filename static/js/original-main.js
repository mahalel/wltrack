// Set up form handling for exercise forms
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
  const addExerciseBtn = document.getElementById("add-exercise");
  if (!addExerciseBtn) return;

  // Add event listener to the "Add Another Exercise" button
  addExerciseBtn.addEventListener("click", function () {
    const exerciseList = document.getElementById("exercise-list");
    const existingExercises = exerciseList.querySelectorAll(".exercise-entry");
    const lastExercise = existingExercises[existingExercises.length - 1];

    // Clone the last exercise entry
    const newExercise = lastExercise.cloneNode(true);

    // Reset form values in the clone
    const inputs = newExercise.querySelectorAll("input");
    inputs.forEach((input) => {
      if (input.name !== "set_start[]" && input.name !== "set_end[]") {
        input.value = "";
      }
    });

    const select = newExercise.querySelector("select");
    select.selectedIndex = 0;

    // Keep only one set range in the new exercise
    const setRangeContainer = newExercise.querySelector(".set-range-container");
    const setRanges = setRangeContainer.querySelectorAll(".set-range");
    for (let i = 1; i < setRanges.length; i++) {
      setRanges[i].remove();
    }

    // Reset the first set range
    const firstSetRange = setRangeContainer.querySelector(".set-range");
    if (firstSetRange) {
      firstSetRange.querySelector('input[name="set_start[]"]').value = "1";
      firstSetRange.querySelector('input[name="set_end[]"]').value = "3";
      firstSetRange.querySelector(".sets-count").textContent = "3";

      // Reset the reps, percentage and weight to default values
      if (firstSetRange.querySelector('input[name="reps[]"]')) {
        firstSetRange.querySelector('input[name="reps[]"]').value = "5";
      }
      if (firstSetRange.querySelector('input[name="percentage[]"]')) {
        firstSetRange.querySelector('input[name="percentage[]"]').value = "75";
      }
      if (firstSetRange.querySelector('input[name="weight[]"]')) {
        firstSetRange.querySelector('input[name="weight[]"]').value = "60";
      }
    }

    // Add the new exercise to the list
    exerciseList.appendChild(newExercise);

    // Re-attach event listeners
    attachSetRangeListeners(newExercise);
    attachExerciseRemoveListeners(newExercise);
    updateSetRangeCounters(newExercise);
  });

  // Add event listeners to existing exercises
  document.querySelectorAll(".exercise-entry").forEach((exercise) => {
    attachSetRangeListeners(exercise);
    attachExerciseRemoveListeners(exercise);
    updateSetRangeCounters(exercise);
  });
}

function attachSetRangeListeners(exerciseEl) {
  // Add another set range
  const addSetRangeBtn = exerciseEl.querySelector(".add-set-range");
  if (addSetRangeBtn) {
    addSetRangeBtn.addEventListener("click", function () {
      const setRangeContainer = exerciseEl.querySelector(
        ".set-range-container",
      );
      const existingSetRanges =
        setRangeContainer.querySelectorAll(".set-range");
      const lastSetRange = existingSetRanges[existingSetRanges.length - 1];

      // Get the last set end value to determine next start
      const lastEndInput = lastSetRange.querySelector(
        'input[name="set_end[]"]',
      );
      const nextSetStart = parseInt(lastEndInput.value) + 1;
      const nextSetEnd = nextSetStart + 2; // Default to 3 sets

      // Clone the last set range
      const newSetRange = lastSetRange.cloneNode(true);

      // Update set range values
      const startInput = newSetRange.querySelector('input[name="set_start[]"]');
      const endInput = newSetRange.querySelector('input[name="set_end[]"]');
      startInput.value = nextSetStart;
      endInput.value = nextSetEnd;

      // Reset other form values in the clone
      const otherInputs = newSetRange.querySelectorAll(
        'input:not([name="set_start[]"]):not([name="set_end[]"])',
      );
      otherInputs.forEach((input) => (input.value = ""));

      // Update sets count
      newSetRange.querySelector(".sets-count").textContent =
        nextSetEnd - nextSetStart + 1;

      // Add event listeners
      attachSetRangeInputHandlers(newSetRange);

      // Add remove button handler
      const removeBtn = newSetRange.querySelector(".remove-set-range");
      if (removeBtn) {
        removeBtn.addEventListener("click", function () {
          newSetRange.remove();
          // Renumber subsequent set ranges
          updateAllSetRanges(exerciseEl);
        });
      }

      // Add the new set range to the container
      setRangeContainer.appendChild(newSetRange);
    });
  }

  // Add handlers to existing set ranges
  exerciseEl.querySelectorAll(".set-range").forEach((setRange) => {
    attachSetRangeInputHandlers(setRange);

    // Add remove button handler
    const removeBtn = setRange.querySelector(".remove-set-range");
    if (removeBtn) {
      removeBtn.addEventListener("click", function () {
        // Only remove if there's more than one set range
        const container = setRange.closest(".set-range-container");
        if (container.querySelectorAll(".set-range").length > 1) {
          setRange.remove();
          // Renumber subsequent set ranges
          updateAllSetRanges(exerciseEl);
        } else {
          alert("You must have at least one set range.");
        }
      });
    }
  });
}

function attachExerciseRemoveListeners(exerciseEl) {
  const removeBtn = exerciseEl.querySelector(".remove-exercise");
  if (!removeBtn) return;

  removeBtn.addEventListener("click", function () {
    const exerciseList = document.getElementById("exercise-list");
    const exercises = exerciseList.querySelectorAll(".exercise-entry");

    // Don't remove if it's the only exercise
    if (exercises.length > 1) {
      exerciseEl.remove();
    } else {
      alert("You must have at least one exercise in your workout.");
    }
  });
}

// Add a new function to handle set range input changes
function attachSetRangeInputHandlers(setRangeEl) {
  const startInput = setRangeEl.querySelector('input[name="set_start[]"]');
  const endInput = setRangeEl.querySelector('input[name="set_end[]"]');
  const countEl = setRangeEl.querySelector(".sets-count");

  // Update sets count when start or end changes
  const updateCount = function () {
    const start = parseInt(startInput.value) || 1;
    const end = parseInt(endInput.value) || start;

    // Ensure end is not less than start
    if (end < start) {
      endInput.value = start;
    }

    // Calculate and update count
    const count =
      (parseInt(endInput.value) || start) -
      (parseInt(startInput.value) || 1) +
      1;
    countEl.textContent = count;
  };

  startInput.addEventListener("input", updateCount);
  endInput.addEventListener("input", updateCount);

  // Handle percentage to weight calculation
  const percentageInput = setRangeEl.querySelector(".percentage-input");
  const weightInput = setRangeEl.querySelector('input[name="weight[]"]');

  if (percentageInput && weightInput) {
    percentageInput.addEventListener("input", function () {
      // Try to find the exercise ID to get the 1RM
      const exerciseEl = setRangeEl.closest(".exercise-entry");
      const exerciseSelect = exerciseEl.querySelector(
        'select[name="exercise_id[]"]',
      );

      if (exerciseSelect && exerciseSelect.value) {
        // Make an AJAX call to get the current 1RM
        fetch(`/api/exercises/${exerciseSelect.value}/1rm`)
          .then((response) => {
            if (!response.ok) throw new Error("Failed to fetch 1RM");
            return response.text();
          })
          .then((html) => {
            // Extract the 1RM value from the response
            const tempDiv = document.createElement("div");
            tempDiv.innerHTML = html;
            const oneRMText =
              tempDiv.querySelector(".font-medium")?.textContent;

            if (oneRMText) {
              // Extract numeric value from "XX kg"
              const oneRM = parseFloat(oneRMText.replace(/[^\d.]/g, ""));

              if (!isNaN(oneRM) && oneRM > 0) {
                const percentage = parseFloat(percentageInput.value) || 0;
                const calculatedWeight = (percentage / 100) * oneRM;

                // Round to nearest 2.5kg
                const roundedWeight = Math.round(calculatedWeight / 2.5) * 2.5;
                weightInput.value = roundedWeight;
              }
            }
          })
          .catch((error) => {
            console.error("Error fetching 1RM:", error);
          });
      }
    });
  }
}

// Function to update counters in all set ranges
function updateSetRangeCounters(exerciseEl) {
  exerciseEl.querySelectorAll(".set-range").forEach((setRange) => {
    const startInput = setRange.querySelector('input[name="set_start[]"]');
    const endInput = setRange.querySelector('input[name="set_end[]"]');
    const countEl = setRange.querySelector(".sets-count");

    if (startInput && endInput && countEl) {
      const count =
        (parseInt(endInput.value) || 1) - (parseInt(startInput.value) || 1) + 1;
      countEl.textContent = count;
    }
  });
}

// Function to update all set ranges (renumbering)
function updateAllSetRanges(exerciseEl) {
  const setRanges = exerciseEl.querySelectorAll(".set-range");
  let nextSetStart = 1;

  setRanges.forEach((setRange, index) => {
    const startInput = setRange.querySelector('input[name="set_start[]"]');
    const endInput = setRange.querySelector('input[name="set_end[]"]');

    if (startInput && endInput) {
      // If it's the first range, it should start at 1
      if (index === 0) {
        startInput.value = 1;
      } else {
        startInput.value = nextSetStart;
      }

      // Calculate the range for this set
      const setsInRange =
        parseInt(endInput.value) - parseInt(startInput.value) + 1;
      endInput.value = parseInt(startInput.value) + setsInRange - 1;

      // Update next start value
      nextSetStart = parseInt(endInput.value) + 1;

      // Update the counter
      const countEl = setRange.querySelector(".sets-count");
      if (countEl) {
        countEl.textContent = setsInRange;
      }
    }
  });
}

// Set up Chart.js visualizations
function setupCharts() {
  // Progress chart on the home page
  setupHomePageChart();

  // Exercise detail page chart
  setupExerciseDetailChart();

  // Make sure forms with hx-put use the right method
  setupHtmxForms();
}

// Fix HTMX form handling for PUT requests
function setupHtmxForms() {
  // Find all forms with hx-put attribute and ensure they use the correct method
  document.querySelectorAll("form[hx-put]").forEach((form) => {
    form.addEventListener("submit", function (e) {
      // Prevent default submission
      e.preventDefault();

      // Get the URL from the hx-put attribute
      const url = this.getAttribute("hx-put");
      const target = this.getAttribute("hx-target");
      const swap = this.getAttribute("hx-swap");

      // Get form data
      const formData = new FormData(this);

      // Make an HTMX request
      htmx.ajax("PUT", url, {
        target: target,
        swap: swap,
        values: formData,
        headers: {
          "X-Requested-With": "XMLHttpRequest",
        },
      });
    });
  });

  // Add input step customization
  document
    .querySelectorAll(
      'input[type="number"][name="one_rep_max"], input[type="number"][name="weight[]"]',
    )
    .forEach((input) => {
      // Ensure arrow key increments are in steps of 5
      input.addEventListener("keydown", function (e) {
        if (e.key === "ArrowUp") {
          e.preventDefault();
          this.value = (parseFloat(this.value || 0) + 5).toString();
        } else if (e.key === "ArrowDown") {
          e.preventDefault();
          const newValue = Math.max(0, parseFloat(this.value || 0) - 5);
          this.value = newValue.toString();
        }
      });
    });
}

function setupHomePageChart() {
  const chartCanvas = document.getElementById("progress-chart");
  const exerciseSelector = document.getElementById("chart-exercise-selector");
  const noDataMessage = document.getElementById("chart-no-data");

  if (!chartCanvas || !exerciseSelector) return;

  let chart = null;

  // Function to load data and update chart
  function updateChart() {
    const exerciseId = exerciseSelector.value;
    if (!exerciseId) return;

    fetch(`/api/exercises/${exerciseId}/history`)
      .then((response) => response.json())
      .then((data) => {
        // Check if data is null, undefined or empty
        if (!data || !Array.isArray(data) || data.length === 0) {
          if (noDataMessage) noDataMessage.classList.remove("hidden");
          if (chartCanvas) chartCanvas.classList.add("hidden");
          return;
        }

        if (noDataMessage) noDataMessage.classList.add("hidden");
        if (chartCanvas) chartCanvas.classList.remove("hidden");

        // Process and sort data for the chart
        // Create an array of objects with date and metrics
        let processedData = data.map(d => {
          // Make sure weights array exists and has elements
          if (!d.weights || !Array.isArray(d.weights) || d.weights.length === 0) {
            return { date: d.date || 'Unknown', avgWeight: 0 };
          }
          return {
            date: d.date || 'Unknown',
            avgWeight: parseFloat((d.weights.reduce((a, b) => a + b, 0) / d.weights.length).toFixed(1)),
          };
        });
      
        // Sort by date (assuming format "Week W, MMM")
        processedData.sort((a, b) => {
          // Extract week number
          const weekA = parseInt(a.date.match(/Week (\d+)/)?.[1] || '0');
          const weekB = parseInt(b.date.match(/Week (\d+)/)?.[1] || '0');
        
          // Extract month (Jan, Feb, etc)
          const monthA = a.date.split(', ')[1];
          const monthB = b.date.split(', ')[1];
        
          // First compare by month (rough approximation)
          const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
          const monthDiff = months.indexOf(monthA) - months.indexOf(monthB);
        
          if (monthDiff !== 0) return monthDiff;
        
          // If same month, compare by week
          return weekA - weekB;
        });
      
        // Extract sorted values for the chart
        const labels = processedData.map(d => d.date);
        const weights = processedData.map(d => d.avgWeight);

        // Create or update chart
        if (chart) {
          chart.data.labels = labels;
          chart.data.datasets[0].data = weights;
          chart.update();
        } else {
          try {
            chart = new window.Chart(chartCanvas, {
              type: "line",
              data: {
                labels: labels && labels.length > 0 ? labels : ['No Data'],
                datasets: [
                  {
                    label: "Average Weight",
                    data: weights && weights.length > 0 ? weights : [0],
                  backgroundColor: "rgba(59, 130, 246, 0.2)",
                  borderColor: "rgba(59, 130, 246, 1)",
                  borderWidth: 2,
                  tension: 0.3,
                  pointRadius: 4,
                  pointHoverRadius: 6,
                },
              ],
            },
            options: {
              responsive: true,
              scales: {
                y: {
                  beginAtZero: false,
                  title: {
                    display: true,
                    text: "Weight (kg)",
                  },
                },
                x: {
                  title: {
                    display: true,
                    text: "Date",
                  },
                  ticks: {
                    maxRotation: 45,
                    minRotation: 45,
                    autoSkip: true,
                    maxTicksLimit: 8,
                    font: {
                      size: 10
                    }
                  }
                },
              },
            },
          });
        } catch (e) {
          console.error("Error creating chart:", e);
          if (noDataMessage) {
            noDataMessage.textContent = "Error creating chart";
            noDataMessage.classList.remove("hidden");
          }
          if (chartCanvas) chartCanvas.classList.add("hidden");
        }
        }
      })
      .catch((error) => {
        console.error("Error fetching exercise data:", error);
        if (noDataMessage) {
          noDataMessage.textContent = "Error loading chart data";
          noDataMessage.classList.remove("hidden");
        }
        if (chartCanvas) chartCanvas.classList.add("hidden");
        // If chart already exists, destroy it to prevent further errors
        if (chart) {
          chart.destroy();
          chart = null;
        }
      });
  }

  // Update chart when exercise changes
  if (exerciseSelector) {
    exerciseSelector.addEventListener("change", updateChart);

    // Initial chart load
    if (exerciseSelector.value) {
      updateChart();
    }
  }
}

// Function to set step to 5kg for all weight inputs
function setupWeightInputs() {
  // Set all weight inputs to have 5kg steps
  const setWeightAttributes = () => {
    document.querySelectorAll('input[name="weight[]"]').forEach((input) => {
      input.setAttribute("step", "5");
      input.addEventListener("change", function () {
        // Round to nearest 5
        const value = parseFloat(this.value);
        if (!isNaN(value)) {
          const rounded = Math.round(value / 5) * 5;
          this.value = rounded;
        }
      });
    });
  };

  // Run once on page load
  setWeightAttributes();

  // Also watch for dynamically added elements
  const observer = new MutationObserver(function (mutations) {
    mutations.forEach(function (mutation) {
      if (mutation.addedNodes.length) {
        setWeightAttributes();
      }
    });
  });

  // Start observing the document
  observer.observe(document.body, { childList: true, subtree: true });
}

function setupExerciseDetailChart() {
  const chartCanvas = document.getElementById("exercise-progress-chart");
  if (!chartCanvas) return;

  // Get exercise ID from URL
  const path = window.location.pathname;
  const match = path.match(/\/exercises\/(\d+)/);
  if (!match) return;

  const exerciseId = match[1];
  const noDataMessage = document.getElementById("chart-no-data");

  fetch(`/api/exercises/${exerciseId}/history`)
    .then((response) => response.json())
    .then((data) => {
      if (data.length === 0) {
        if (noDataMessage) noDataMessage.classList.remove("hidden");
        if (chartCanvas) chartCanvas.classList.add("hidden");
        return;
      }

      if (noDataMessage) noDataMessage.classList.add("hidden");
      if (chartCanvas) chartCanvas.classList.remove("hidden");

      // Process and sort data for the chart
      // Create an array of objects with date and metrics
      let processedData = data.map(d => ({
        date: d.date,
        avgWeight: parseFloat((d.weights.reduce((a, b) => a + b, 0) / d.weights.length).toFixed(1)),
        maxWeight: parseFloat(Math.max(...d.weights).toFixed(1))
      }));
      
      // Sort by date (assuming format "Week W, MMM")
      processedData.sort((a, b) => {
        // Extract week number
        const weekA = parseInt(a.date.match(/Week (\d+)/)?.[1] || '0');
        const weekB = parseInt(b.date.match(/Week (\d+)/)?.[1] || '0');
        
        // Extract month (Jan, Feb, etc)
        const monthA = a.date.split(', ')[1];
        const monthB = b.date.split(', ')[1];
        
        // First compare by month (rough approximation)
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        const monthDiff = months.indexOf(monthA) - months.indexOf(monthB);
        
        if (monthDiff !== 0) return monthDiff;
        
        // If same month, compare by week
        return weekA - weekB;
      });
      
      // Reverse order to show oldest first on the X-axis (left to right)
      processedData.reverse();
      
      // Limit to 12 most recent entries for better readability
      if (processedData.length > 12) {
        processedData = processedData.slice(processedData.length - 12);
      }
      
      // Extract sorted values for the chart
      const labels = processedData.map(d => d.date);
      const avgWeights = processedData.map(d => d.avgWeight);
      const maxWeights = processedData.map(d => d.maxWeight);

      // Create chart
      const chart = new window.Chart(chartCanvas, {
        type: "line",
        data: {
          labels: labels,
          datasets: [
            {
              label: "Average Weight",
              data: avgWeights,
              backgroundColor: "rgba(59, 130, 246, 0.2)",
              borderColor: "rgba(59, 130, 246, 1)",
              borderWidth: 2,
              tension: 0.3,
              pointRadius: 4,
              pointHoverRadius: 6,
            },
            {
              label: "Max Weight",
              data: maxWeights,
              backgroundColor: "rgba(220, 38, 38, 0.2)",
              borderColor: "rgba(220, 38, 38, 1)",
              borderWidth: 2,
              tension: 0.3,
              pointRadius: 4,
              pointHoverRadius: 6,
            },
          ],
        },
        options: {
          responsive: true,
          scales: {
            y: {
              beginAtZero: false,
              title: {
                display: true,
                text: "Weight (kg)",
              },
            },
            x: {
              title: {
                display: true,
                text: "Date",
              },
              ticks: {
                maxRotation: 45,
                minRotation: 45,
                autoSkip: true,
                maxTicksLimit: 8,
                font: {
                  size: 10
                }
              }
            },
          },
        },
      });
    })
    .catch((error) => {
      console.error("Error fetching exercise data:", error);
      if (noDataMessage) {
        noDataMessage.textContent = "Error loading chart data";
        noDataMessage.classList.remove("hidden");
      }
      if (chartCanvas) chartCanvas.classList.add("hidden");
    });
}
