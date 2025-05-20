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
    attachAddSetRangeListeners(exerciseDiv);
    
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
  attachAddSetRangeListeners(document);
  updateAllSetRanges();
  updateAllSetRangeCounters();
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
      
      // Update the set counters for this exercise
      updateSetRangeCounters(exerciseDiv);
    });
  });
}

function attachAddSetRangeListeners(container) {
  // Add new set range to an exercise
  container.querySelectorAll(".add-set-range").forEach(button => {
    // Remove existing event listeners to prevent duplication
    const newButton = button.cloneNode(true);
    button.parentNode.replaceChild(newButton, button);
    
    newButton.addEventListener("click", function() {
      const exerciseDiv = this.closest(".exercise-entry");
      const setRangeContainer = exerciseDiv.querySelector(".set-range-container");
      
      // Clone the first set range as a template
      const firstSetRange = setRangeContainer.querySelector(".set-range");
      const newSetRange = firstSetRange.cloneNode(true);
      
      // Find the previous set range's end value
      const setRanges = setRangeContainer.querySelectorAll(".set-range");
      const previousSetRange = setRanges[setRanges.length - 1];
      const previousEndInput = previousSetRange.querySelector("input[name='set_end[]']");
      const previousEndValue = previousEndInput ? parseInt(previousEndInput.value) || 0 : 0;
      
      // Calculate new start and end values
      const newStartValue = previousEndValue + 1;
      const newEndValue = newStartValue + 2; // Default to 3 sets per range
      
      // Set input values appropriately
      newSetRange.querySelectorAll("input").forEach(input => {
        if (input.name === "set_start[]") {
          input.value = newStartValue.toString();
        } else if (input.name === "set_end[]") {
          input.value = newEndValue.toString();
        } else {
          input.value = "";
        }
      });
      
      // Update the sets count
      const setsCountElement = newSetRange.querySelector(".sets-count");
      if (setsCountElement) {
        setsCountElement.textContent = (newEndValue - newStartValue + 1).toString();
      }
      
      // Add the new set range to the container
      setRangeContainer.appendChild(newSetRange);
      
      // Attach remove listener to the new set range
      const removeButton = newSetRange.querySelector(".remove-set-range");
      if (removeButton) {
        removeButton.addEventListener("click", function() {
          // Only remove if there's more than one set range
          if (setRangeContainer.querySelectorAll(".set-range").length > 1) {
            newSetRange.remove();
          }
        });
      }
      
      // Attach input handlers to the new set range
      attachSetRangeInputHandlers(newSetRange);
      
      // Ensure the set count is correct
      updateSetRangeCounters(exerciseDiv);
    });
  });
  
  // Add remove listeners to existing set ranges
  container.querySelectorAll(".remove-set-range").forEach(button => {
    // Remove existing event listeners to prevent duplication
    const newButton = button.cloneNode(true);
    button.parentNode.replaceChild(newButton, button);
    
    newButton.addEventListener("click", function() {
      const setRange = this.closest(".set-range");
      const setRangeContainer = this.closest(".set-range-container");
      
      // Only remove if there's more than one set range
      if (setRangeContainer.querySelectorAll(".set-range").length > 1) {
        // Get all set ranges before removal
        const allSetRanges = Array.from(setRangeContainer.querySelectorAll(".set-range"));
        const currentIndex = allSetRanges.indexOf(setRange);
        
        // Get current range values before removal
        const startInput = setRange.querySelector("input[name='set_start[]']");
        const endInput = setRange.querySelector("input[name='set_end[]']");
        const startValue = parseInt(startInput?.value) || 0;
        const endValue = parseInt(endInput?.value) || 0;
        const rangeDiff = (endValue - startValue) + 1;
        
        // Remove the current set range
        setRange.remove();
        
        // Adjust subsequent set ranges to maintain continuity
        if (currentIndex >= 0 && currentIndex < allSetRanges.length - 1) {
          for (let i = currentIndex + 1; i < allSetRanges.length; i++) {
            const nextSetRange = allSetRanges[i];
            const nextStartInput = nextSetRange.querySelector("input[name='set_start[]']");
            const nextEndInput = nextSetRange.querySelector("input[name='set_end[]']");
            
            if (nextStartInput && nextEndInput) {
              const nextStart = parseInt(nextStartInput.value) || 0;
              const nextEnd = parseInt(nextEndInput.value) || 0;
              
              // Adjust the range by the size of the removed range
              nextStartInput.value = Math.max(1, nextStart - rangeDiff);
              nextEndInput.value = Math.max(1, nextEnd - rangeDiff);
              
              // Update the sets count display
              const setsCountElement = nextSetRange.querySelector(".sets-count");
              if (setsCountElement) {
                const newCount = parseInt(nextEndInput.value) - parseInt(nextStartInput.value) + 1;
                setsCountElement.textContent = newCount.toString();
              }
            }
          }
        }
        
        // Update set counters after removal
        updateSetRangeCounters(setRangeContainer.closest(".exercise-entry") || document);
      }
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

// Corrected and with logging
function calculateAndUpdateWeightForSet(setElement) {
  console.log("calculateAndUpdateWeightForSet called for setElement:", setElement);

  if (!setElement) {
    console.error("calculateAndUpdateWeightForSet: setElement is null or undefined. This should be a .set-range div.");
    return;
  }

  const exerciseItem = setElement.closest(".exercise-entry"); // Corrected class
  if (!exerciseItem) {
    console.error("calculateAndUpdateWeightForSet: .exercise-entry not found for setElement:", setElement);
    return;
  }
  console.log("Found exerciseItem (.exercise-entry):", exerciseItem);

  // Corrected selectors based on workouts.templ
  const percentageInput = setElement.querySelector("input[name='percentage[]']"); 
  const exerciseSelect = exerciseItem.querySelector("select[name='exercise_id[]']"); 
  const weightInput = setElement.querySelector("input[name='weight[]']");

  if (!percentageInput) {
    console.warn("calculateAndUpdateWeightForSet: percentageInput (input[name='percentage[]']) not found in setElement:", setElement);
    return; 
  }
  console.log("Found percentageInput:", percentageInput, "Value:", percentageInput.value);

  if (!exerciseSelect) {
    console.warn("calculateAndUpdateWeightForSet: exerciseSelect (select[name='exercise_id[]']) not found in exerciseItem:", exerciseItem);
    return;
  }
  console.log("Found exerciseSelect:", exerciseSelect, "Value:", exerciseSelect.value);
  
  if (!weightInput) {
    console.warn("calculateAndUpdateWeightForSet: weightInput (input[name='weight[]']) not found in setElement:", setElement);
    return;
  }
  console.log("Found weightInput:", weightInput);

  const percentage = parseFloat(percentageInput.value);
  const exerciseId = exerciseSelect.value;

  console.log(`Percentage: ${percentage} (Raw: '${percentageInput.value}'), Exercise ID: ${exerciseId}`);

  if (!exerciseId) {
    console.log("No exercise selected. Not calculating weight.");
    // weightInput.value = ""; // Keep manual value or clear?
    // weightInput.placeholder = "Select exercise";
    return;
  }

  // Do not proceed if percentage is not a positive number.
  // This allows manual weight entry if percentage is empty or zero.
  if (isNaN(percentage) || percentage <= 0) {
    console.log("Percentage is not a positive number. Not auto-calculating weight. Manual input is allowed.");
    // weightInput.placeholder = "Enter % or weight"; // Update placeholder if needed
    return;
  }

  weightInput.placeholder = "Calculating...";
  console.log(`Fetching 1RM for exercise ID: ${exerciseId}`);

  fetch(`/api/exercises/${exerciseId}/1rm`)
    .then(response => {
      console.log("Fetch response status:", response.status);
      if (!response.ok) {
        // If 1RM is not found (e.g. 404), it's not necessarily a hard error for this function.
        // It just means we can't auto-calculate.
        if (response.status === 404) {
            console.log(`No 1RM record found for exercise ID: ${exerciseId}. Manual weight input required.`);
            weightInput.placeholder = "No 1RM on file";
        } else {
            console.error(`HTTP error fetching 1RM! status: ${response.status}, for exercise ID: ${exerciseId}`);
            weightInput.placeholder = "Error fetching 1RM";
        }
        // Clear the value only if we intended to auto-calculate but failed,
        // otherwise, a manually entered value might be present.
        // For now, let's not clear it here, allowing manual override to persist.
        // weightInput.value = ""; 
        return null; // Signal to skip further processing
      }
      return response.text();
    })
    .then(html => {
      if (html === null) return; // Skip if fetch indicated no 1RM or error

      console.log("Received 1RM HTML:", html);
      const tempDiv = document.createElement('div');
      tempDiv.innerHTML = html;
      const oneRmValueEl = tempDiv.querySelector('[data-one-rm-value]');
      
      if (!oneRmValueEl) {
        console.warn("1RM data attribute [data-one-rm-value] not found in response for exercise ID:", exerciseId);
        weightInput.placeholder = "1RM data missing";
        // weightInput.value = "";
        return;
      }
      
      const oneRmString = oneRmValueEl.getAttribute('data-one-rm-value');
      console.log("1RM string from attribute:", oneRmString);
      const oneRm = parseFloat(oneRmString);
      
      if (isNaN(oneRm) || oneRm <= 0) {
        console.warn("Invalid or zero 1RM value parsed:", oneRm, "(from string:", oneRmString, ") for exercise ID:", exerciseId);
        weightInput.placeholder = "Invalid 1RM data";
        // weightInput.value = "";
        return;
      }

      console.log(`Calculating weight: (${percentage} / 100) * ${oneRm}`);
      const calculatedWeight = (percentage / 100) * oneRm;
      const roundedWeight = Math.round(calculatedWeight * 2) / 2;
      console.log("Calculated weight:", calculatedWeight, "Rounded:", roundedWeight);
      
      // Only update if the new value is different, to avoid disrupting manual input if not necessary
      // However, if percentage changes, we *should* update.
      // The condition `if (isNaN(percentage) || percentage <= 0)` above handles when not to auto-calculate.
      weightInput.value = roundedWeight;
      weightInput.dispatchEvent(new Event('input', { bubbles: true })); // For sliders etc.
      weightInput.placeholder = ""; // Clear placeholder as we have a value
      console.log("Weight input updated to:", roundedWeight);
    })
    .catch(error => {
      console.error("Catch block: Error fetching 1RM for exercise ID " + exerciseId + ":", error);
      weightInput.placeholder = "Error calc weight";
      // weightInput.value = "";
    });
}

function attachSetRangeInputHandlers(container) {
  console.log("attachSetRangeInputHandlers called for container:", container);
  // Corrected input selectors based on workouts.templ name attributes
  const inputSelectors = [
    "input[name='weight[]']", 
    "input[name='reps[]']",   
    "input[name='set_start[]']",
    "input[name='set_end[]']",
    // "select[name='rpe[]']", // RPE not directly used in weight calc, but good to re-attach listeners
    "input[type='range']", // For sliders if any are dynamically added with sets
    "input[name='percentage[]']"
  ];
  
  inputSelectors.forEach(selector => {
    container.querySelectorAll(selector).forEach(input => {
      const newInput = input.cloneNode(true);
      input.parentNode.replaceChild(newInput, input);
      console.log("Attached listener for selector:", selector, "to element:", newInput);
      
      // Range slider logic (if applicable to these inputs, mostly for reps/weight if sliders exist)
      if (newInput.type === "range") {
        const counter = newInput.previousElementSibling;
        const actualInputName = newInput.dataset.target;
        if (actualInputName) {
            const actualInput = container.querySelector(`input[name='${actualInputName}'], select[name='${actualInputName}']`);
            if (actualInput) {
                newInput.addEventListener("input", function () {
                    if (counter) counter.textContent = this.value;
                    actualInput.value = this.value;
                    actualInput.dispatchEvent(new Event('input', { bubbles: true }));
                });
                actualInput.addEventListener("input", function () {
                    newInput.value = this.value;
                    if (counter) counter.textContent = this.value;
                });
                if (actualInput.value) {
                    newInput.value = actualInput.value;
                    if (counter) counter.textContent = actualInput.value;
                } else if (newInput.value) {
                    actualInput.value = newInput.value;
                    if (counter) counter.textContent = newInput.value;
                    actualInput.dispatchEvent(new Event('input', { bubbles: true }));
                }
            } else if (counter && newInput.value) {
                 counter.textContent = newInput.value;
            }
        } else if (counter && newInput.value) { // Fallback if data-target is missing
            counter.textContent = newInput.value;
        }
      }
      
      // Update sliders if this input (weight, reps) has a corresponding slider
      if (newInput.name && (newInput.name.includes('weight[]') || newInput.name.includes('reps[]'))) {
        newInput.addEventListener("input", function () {
          const exerciseItem = this.closest(".exercise-entry"); // Corrected class
          if (exerciseItem) {
              const rangeSlider = exerciseItem.querySelector(`input[type='range'][data-target='${this.name}']`);
              if (rangeSlider) {
                  rangeSlider.value = this.value;
                  const counter = rangeSlider.previousElementSibling;
                  if (counter) counter.textContent = this.value;
              }
          }
        });
      }
      
      // For percentage and reps inputs, trigger weight calculation
      if (newInput.name) {
        // Add input event listeners for set_start and set_end inputs
        if (newInput.name === 'set_start[]' || newInput.name === 'set_end[]') {
          newInput.addEventListener("input", function () {
            console.log(`Input event on ${this.name}, value: ${this.value}`);
            const setElement = this.closest(".set-range"); // Corrected class
            if (setElement) {
              // Get the start and end inputs
              const startInput = setElement.querySelector("input[name='set_start[]']");
              const endInput = setElement.querySelector("input[name='set_end[]']");
            
              if (startInput && endInput) {
                const startValue = parseInt(startInput.value) || 1;
                const endValue = parseInt(endInput.value) || 1;
              
                // Enforce relationship: start should never be greater than end
                if (this.name === 'set_start[]' && startValue > endValue) {
                  endInput.value = startValue;
                }
              
                // Enforce relationship: end should never be less than start
                if (this.name === 'set_end[]' && endValue < startValue) {
                  startInput.value = endValue;
                }
              }
            
              calculateAndUpdateWeightForSet(setElement);
              // Update set counters if set_start or set_end changed
              updateSetRangeCounters(setElement.closest(".exercise-entry") || document);
            } else {
              console.warn("Could not find .set-range parent for input:", this);
            }
          });
        } else if (newInput.name.includes('percentage[]') || newInput.name.includes('reps[]')) {
          newInput.addEventListener("input", function () {
            console.log(`Input event on ${this.name}, value: ${this.value}`);
            const setElement = this.closest(".set-range"); // Corrected class
            if (setElement) {
              calculateAndUpdateWeightForSet(setElement);
            } else {
              console.warn("Could not find .set-range parent for input:", this);
            }
          });
        }
      }
    });
  });

  // Handling for exercise selection dropdowns
  container.querySelectorAll("select[name='exercise_id[]']").forEach(exerciseSelect => { // Corrected selector
    const newExerciseSelect = exerciseSelect.cloneNode(true);
    exerciseSelect.parentNode.replaceChild(newExerciseSelect, exerciseSelect);
    console.log("Attached change listener to exercise select:", newExerciseSelect);
    
    newExerciseSelect.addEventListener("change", function() {
      console.log(`Change event on exercise select, new value: ${this.value}`);
      const exerciseItem = this.closest(".exercise-entry"); // Corrected class
      if (exerciseItem) {
        exerciseItem.querySelectorAll(".set-range").forEach(setElement => { // Corrected class
          calculateAndUpdateWeightForSet(setElement);
        });
      } else {
         console.warn("Could not find .exercise-entry parent for select:", this);
      }
    });
  });

  // After attaching all handlers, try an initial calculation for all sets in the container.
  // This helps if the form loads with pre-filled percentage values or selected exercises.
  if (container.matches || container.querySelectorAll) { // Check if container is a valid element/document fragment
      const sets = (typeof container.querySelectorAll === 'function') ? container.querySelectorAll(".set-range") : [];
      console.log("Attempting initial calculation for sets in container:", sets.length);
      sets.forEach(set => {
          calculateAndUpdateWeightForSet(set);
      });
  }
}

function updateSetRangeCounters(exerciseDiv) {
  if (!exerciseDiv) return;
  
  // Get all set ranges within this exercise
  const setRanges = exerciseDiv.querySelectorAll(".set-range");
  
  // Update the count for each set range
  setRanges.forEach(setRange => {
    const startInput = setRange.querySelector("input[name='set_start[]']");
    const endInput = setRange.querySelector("input[name='set_end[]']");
    const setsCountElement = setRange.querySelector(".sets-count");
    
    if (startInput && endInput && setsCountElement) {
      const start = parseInt(startInput.value) || 0;
      const end = parseInt(endInput.value) || 0;
      
      // Calculate total sets (inclusive range)
      const totalSets = end >= start ? (end - start + 1) : 0;
      
      // Update the displayed count
      setsCountElement.textContent = totalSets;
    }
  });
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

function updateAllSetRangeCounters() {
  // Get all exercise entries
  const exerciseEntries = document.querySelectorAll(".exercise-entry");
  
  // Update set counts for each exercise entry
  exerciseEntries.forEach(exerciseEntry => {
    updateSetRangeCounters(exerciseEntry);
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
                    return context.dataset.label + ': ' + context.raw + ' kg';
                  }
                }
              }
            },
          scales: {
            y: {
              title: {
                display: true,
                text: 'Weight (kg)'
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
  
  // Note: We no longer need to attach percentage event listeners here
  // They're now handled in attachSetRangeInputHandlers for all inputs
  // including dynamically added ones
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
                return label + ': ' + value + ' kg';
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
            text: 'Weight (kg)'
          },
          beginAtZero: false
        },
        y1: {
          position: 'right',
          title: {
            display: true,
            text: 'Volume (kg)'
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