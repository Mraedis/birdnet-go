// handlers.go: This file contains the request handlers for the web server.
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/datastore"
)

// NoteWithSpectrogram extends the Note model with a spectrogram path.
type NoteWithSpectrogram struct {
	datastore.Note
	Spectrogram string
}

// topBirdsHandler handles requests for the top bird sightings.
// It retrieves data based on the specified date and minimum confidence,
// then renders it using the 'birdsTableHTML' template.
func (h *Handlers) TopBirds(c echo.Context) error {
	// Retrieving query parameters
	selectedDate := c.QueryParam("date")
	if selectedDate == "" {
		selectedDate = getCurrentDate()
	}

	// Parse the selected date
	parsedDate, err := time.Parse("2006-01-02", selectedDate)
	if err != nil {
		return h.NewHandlerError(err, "Invalid date format", http.StatusBadRequest)
	}

	minConfidenceStr := c.QueryParam("minConfidence")
	minConfidence, err := strconv.ParseFloat(minConfidenceStr, 64)
	if err != nil {
		minConfidence = 0.0 // Default value on error
	}
	minConfidenceNormalized := minConfidence / 100.0

	// Get top birds data from the database
	notes, err := h.DS.GetTopBirdsData(selectedDate, minConfidenceNormalized)
	if err != nil {
		return h.NewHandlerError(err, "Failed to get top birds data", http.StatusInternalServerError)
	}

	// Process notes with additional data such as hourly occurrences and total detections
	notesWithIndex, err := h.ProcessNotes(notes, selectedDate, minConfidenceNormalized)
	if err != nil {
		return h.NewHandlerError(err, "Failed to process notes", http.StatusInternalServerError)
	}

	// Sorting the notes by total detections in descending order
	sort.Slice(notesWithIndex, func(i, j int) bool {
		return notesWithIndex[i].TotalDetections > notesWithIndex[j].TotalDetections
	})

	// Creating a slice with hours from 0 to 23
	hours := makeHoursSlice()

	// Get sunrise time
	sunrise, err := h.SunCalc.GetSunriseTime(parsedDate)
	if err != nil {
		return h.NewHandlerError(err, "Failed to get sunrise time", http.StatusInternalServerError)
	}

	// Get sunset time
	sunset, err := h.SunCalc.GetSunsetTime(parsedDate)
	if err != nil {
		return h.NewHandlerError(err, "Failed to get sunset time", http.StatusInternalServerError)
	}

	sunriseHour := closestHour(sunrise)
	sunsetHour := closestHour(sunset)

	// Preparing data for rendering in the template
	data := struct {
		NotesWithIndex    []NoteWithIndex
		Hours             []int
		SelectedDate      string
		DashboardSettings *conf.Dashboard
		Sunrise           int
		Sunset            int
	}{
		NotesWithIndex:    notesWithIndex,
		Hours:             hours,
		SelectedDate:      selectedDate,
		DashboardSettings: h.DashboardSettings,
		Sunrise:           sunriseHour,
		Sunset:            sunsetHour,
	}

	// Render the birdsTableHTML template with the data
	return c.Render(http.StatusOK, "birdsTableHTML", data)
}

func closestHour(t time.Time) int {
	hour := t.Hour()
	minute := t.Minute()

	if minute >= 30 {
		hour++
	}

	return hour % 24
}

// Additional helper functions (processNotes, makeHoursSlice, updateClipNames, etc.) go here...
func (h *Handlers) ProcessNotes(notes []datastore.Note, selectedDate string, minConfidenceNormalized float64) ([]NoteWithIndex, error) {
	startTime := time.Now()
	notesWithIndex := make([]NoteWithIndex, 0, len(notes))
	for _, note := range notes {
		//noteStartTime := time.Now() // Start timing for this note

		hourlyCounts, err := h.DS.GetHourlyOccurrences(selectedDate, note.CommonName, minConfidenceNormalized)
		if err != nil {
			return nil, err // Return error to be handled by the caller
		}

		//log.Printf("Time to fetch hourly occurrences for note %s: %v", note.CommonName, time.Since(noteStartTime)) // Print the time taken for this part

		totalDetections := sumHourlyCounts(hourlyCounts)

		notesWithIndex = append(notesWithIndex, NoteWithIndex{
			Note:            note,
			HourlyCounts:    hourlyCounts,
			TotalDetections: totalDetections,
		})
		//log.Printf("Total time for processing note %s: %v", note.CommonName, time.Since(noteStartTime)) // Print the total time for this note

	}
	log.Printf("Total time for processing all notes: %v", time.Since(startTime)) // Print the total time for the function

	// return notes with hourly counts and total detections
	return notesWithIndex, nil
}

// getAllNotes retrieves all notes from the database.
// It returns the notes in a JSON format.
func (h *Handlers) GetAllNotes(c echo.Context) error {
	notes, err := h.DS.GetAllNotes()
	if err != nil {
		return h.NewHandlerError(err, "Failed to get all notes", http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, notes)
}

// deleteNoteHandler deletes note object from database and its associated audio file
func (h *Handlers) DeleteNote(c echo.Context) error {
	noteID := c.QueryParam("id")
	if noteID == "" {
		return h.NewHandlerError(fmt.Errorf("empty note ID"), "Note ID is required", http.StatusBadRequest)
	}

	// Retrieve the path to the audio file before deleting the note
	clipPath, err := h.DS.GetNoteClipPath(noteID)
	if err != nil {
		return h.NewHandlerError(err, "Failed to retrieve audio clip path", http.StatusInternalServerError)
	}

	// Delete the note from the database
	err = h.DS.Delete(noteID)
	if err != nil {
		return h.NewHandlerError(err, "Failed to delete note", http.StatusInternalServerError)
	}

	// If there's an associated clip, delete the file
	if clipPath != "" {
		err = os.Remove(clipPath)
		if err != nil {
			h.logError(&HandlerError{Err: err, Message: "Failed to delete audio clip", Code: http.StatusInternalServerError})
		} else {
			h.logInfo(fmt.Sprintf("Deleted audio clip: %s", clipPath))
		}
	}

	// Pass this struct to the template or return a success message
	return c.HTML(http.StatusOK, `<div x-data="{ show: true }" x-show="show" x-init="setTimeout(() => show = false, 3000)" class="notification-class">Delete successful!</div>`)
}
