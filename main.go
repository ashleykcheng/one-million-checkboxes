package main

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"net/http"
	"strconv"
)

const NumCheckboxes = 1_000_000

type CheckboxSet struct {
	bits []uint64
}

func NewCheckboxSet() *CheckboxSet {
	return &CheckboxSet{
		bits: make([]uint64, (NumCheckboxes+63)/64),
	}
}

func (cs *CheckboxSet) Toggle(index int) {
	wordIndex := index / 64
	bitIndex := index % 64
	cs.bits[wordIndex] ^= 1 << bitIndex
}

func (cs *CheckboxSet) IsChecked(index int) bool {
	wordIndex := index / 64
	bitIndex := index % 64
	return (cs.bits[wordIndex] & (1 << bitIndex)) != 0
}

func (cs *CheckboxSet) CountChecked() int {
	count := 0
	for _, word := range cs.bits {
		count += bits.OnesCount64(word)
	}
	return count
}

var checkboxes *CheckboxSet

func main() {
	checkboxes = NewCheckboxSet()

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/count", countHandler)
	http.HandleFunc("/state", stateHandler)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Checkbox Manager</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        #checkboxes { display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: 10px; }
        .checkbox { display: flex; align-items: center; }
        #count { margin-top: 20px; font-weight: bold; }
    </style>
</head>
<body>
    <h1>Checkbox Manager</h1>
    <div id="checkboxes"></div>
    <div id="count"></div>

    <script>
        const checkboxesContainer = document.getElementById('checkboxes');
        const countElement = document.getElementById('count');

        function createCheckboxes(num) {
            for (let i = 0; i < num; i++) {
                const checkbox = document.createElement('div');
                checkbox.className = 'checkbox';
                checkbox.innerHTML = '<input type="checkbox" id="cb' + i + '"><label for="cb' + i + '">' + i + '</label>';
                checkbox.querySelector('input').addEventListener('change', (e) => toggleCheckbox(i, e.target.checked));
                checkboxesContainer.appendChild(checkbox);
            }
        }

        function toggleCheckbox(index, checked) {
            fetch('/toggle?index=' + index)
                .then(response => response.text())
                .then(() => updateCount());
        }

        function updateCount() {
            fetch('/count')
                .then(response => response.text())
                .then(count => {
                    countElement.textContent = 'Checked boxes: ' + count;
                });
        }

        function updateState() {
            fetch('/state')
                .then(response => response.json())
                .then(state => {
                    state.forEach((checked, i) => {
                        document.getElementById('cb' + i).checked = checked;
                    });
                    updateCount();
                });
        }

        createCheckboxes(100); // Create 100 checkboxes for demonstration
        updateState();
        setInterval(updateState, 5000); // Update state every 5 seconds
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= NumCheckboxes {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	checkboxes.Toggle(index)
	fmt.Fprintf(w, "Toggled checkbox %d", index)
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	count := checkboxes.CountChecked()
	fmt.Fprintf(w, "%d", count)
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	state := make([]bool, NumCheckboxes)
	for i := 0; i < NumCheckboxes; i++ {
		state[i] = checkboxes.IsChecked(i)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}