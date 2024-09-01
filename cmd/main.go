package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/RedrikShuhartRed/TaskManager/models"
	titleentry "github.com/RedrikShuhartRed/TaskManager/titleEntry"
)

var (
	portContainer *fyne.Container
	actionsScroll *fyne.Container
	serverPort    string
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Verba Task Manager")

	portEntry := titleentry.CreateEntry("Enter server port")
	resultLabel := widget.NewLabel("Result will be displayed here")

	updateResultLabel := func(statusCode int, message string) {
		resultLabel.SetText(fmt.Sprintf("Status Code: %d\nMessage: %s", statusCode, message))
	}

	getBaseURL := func() string {
		return fmt.Sprintf("http://localhost:%s", serverPort)
	}

	titleEntry := titleentry.CreateEntry("Enter task title")
	descriptionEntry := titleentry.CreateEntry("Enter task description")
	dueDateEntry := titleentry.CreateEntry("Enter due date (RFC3339 format)")
	idEntry := titleentry.CreateEntry("Enter task ID")
	updateIDEntry := titleentry.CreateEntry("Enter task ID to update")
	updateTitleEntry := titleentry.CreateEntry("Enter new task title")
	updateDescriptionEntry := titleentry.CreateEntry("Enter new task description")
	updateDueDateEntry := titleentry.CreateEntry("Enter new due date (RFC3339 format)")
	deleteIDEntry := titleentry.CreateEntry("Enter task ID to delete")

	addButton := widget.NewButton("Add Task", func() {
		title := titleEntry.Text
		description := descriptionEntry.Text
		dueDateStr := dueDateEntry.Text

		task := models.Task{
			Title:       title,
			Description: description,
		}

		dueDate, err := time.Parse(time.RFC3339, dueDateStr)
		if err != nil {
			updateResultLabel(http.StatusBadRequest, fmt.Sprintf("Invalid due date format. Error: %v", err))
			return
		}
		task.DueDate = dueDate

		taskData, err := json.Marshal(task)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error marshaling task data: %v", err))
			return
		}

		resp, err := http.Post(getBaseURL()+"/tasks", "application/json", bytes.NewBuffer(taskData))
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			var createdTask models.Task
			err := json.NewDecoder(resp.Body).Decode(&createdTask)
			if err != nil {
				updateResultLabel(resp.StatusCode, fmt.Sprintf("Error decoding response: %v", err))
				return
			}
			updateResultLabel(resp.StatusCode, fmt.Sprintf("Task created successfully!\n\nID: %d\nTitle: %s\nDescription: %s\nDue Date: %s\nCreated At: %s\nUpdated At: %s",
				createdTask.ID,
				createdTask.Title,
				createdTask.Description,
				createdTask.DueDate.Format(time.RFC3339),
				createdTask.CreatedAt.Format(time.RFC3339),
				createdTask.UpdatedAt.Format(time.RFC3339),
			))
		} else {
			updateResultLabel(resp.StatusCode, "Failed to create task")
		}
	})

	getAllTasksButton := widget.NewButton("Get All Tasks", func() {
		resp, err := http.Get(getBaseURL() + "/tasks")
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var tasks []models.Task
			err := json.NewDecoder(resp.Body).Decode(&tasks)
			if err != nil {
				updateResultLabel(resp.StatusCode, fmt.Sprintf("Error decoding response: %v", err))
				return
			}

			var resultText string
			for _, task := range tasks {
				resultText += fmt.Sprintf("ID: %d\nTitle: %s\nDescription: %s\nDue Date: %s\nCreated At: %s\nUpdated At: %s\n\n",
					task.ID,
					task.Title,
					task.Description,
					task.DueDate.Format(time.RFC3339),
					task.CreatedAt.Format(time.RFC3339),
					task.UpdatedAt.Format(time.RFC3339),
				)
			}
			updateResultLabel(resp.StatusCode, resultText)
		} else {
			updateResultLabel(resp.StatusCode, "Failed to retrieve tasks")
		}
	})

	getTaskByIDButton := widget.NewButton("Get Task By ID", func() {
		idStr := idEntry.Text

		resp, err := http.Get(getBaseURL() + "/tasks/" + idStr)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var task models.Task
			err := json.NewDecoder(resp.Body).Decode(&task)
			if err != nil {
				updateResultLabel(resp.StatusCode, fmt.Sprintf("Error decoding response: %v", err))
				return
			}
			updateResultLabel(resp.StatusCode, fmt.Sprintf("Task Details:\n\nID: %d\nTitle: %s\nDescription: %s\nDue Date: %s\nCreated At: %s\nUpdated At: %s",
				task.ID,
				task.Title,
				task.Description,
				task.DueDate.Format(time.RFC3339),
				task.CreatedAt.Format(time.RFC3339),
				task.UpdatedAt.Format(time.RFC3339),
			))
		} else if resp.StatusCode == http.StatusNotFound {
			updateResultLabel(resp.StatusCode, "Task not found")
		} else {
			updateResultLabel(resp.StatusCode, "Failed to retrieve task")
		}
	})

	updateTaskByIDButton := widget.NewButton("Update Task By ID", func() {
		idStr := updateIDEntry.Text
		title := updateTitleEntry.Text
		description := updateDescriptionEntry.Text
		dueDateStr := updateDueDateEntry.Text

		id, err := strconv.Atoi(idStr)
		if err != nil {
			updateResultLabel(http.StatusBadRequest, fmt.Sprintf("Invalid ID format. Error: %v", err))
			return
		}

		dueDate, err := time.Parse(time.RFC3339, dueDateStr)
		if err != nil {
			updateResultLabel(http.StatusBadRequest, fmt.Sprintf("Invalid due date format. Error: %v", err))
			return
		}

		task := models.Task{
			ID:          id,
			Title:       title,
			Description: description,
			DueDate:     dueDate,
		}

		taskData, err := json.Marshal(task)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error marshaling task data: %v", err))
			return
		}

		req, err := http.NewRequest(http.MethodPut, getBaseURL()+"/tasks/"+idStr, bytes.NewBuffer(taskData))
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error creating request: %v", err))
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var updatedTask models.Task
			err := json.NewDecoder(resp.Body).Decode(&updatedTask)
			if err != nil {
				updateResultLabel(resp.StatusCode, fmt.Sprintf("Error decoding response: %v", err))
				return
			}
			updateResultLabel(resp.StatusCode, fmt.Sprintf("Task updated successfully!\n\nID: %d\nTitle: %s\nDescription: %s\nDue Date: %s\nCreated At: %s\nUpdated At: %s",
				updatedTask.ID,
				updatedTask.Title,
				updatedTask.Description,
				updatedTask.DueDate.Format(time.RFC3339),
				updatedTask.CreatedAt.Format(time.RFC3339),
				updatedTask.UpdatedAt.Format(time.RFC3339),
			))
		} else if resp.StatusCode == http.StatusNotFound {
			updateResultLabel(resp.StatusCode, "Task not found")
		} else {
			updateResultLabel(resp.StatusCode, "Failed to update task")
		}
	})

	deleteTaskByIDButton := widget.NewButton("Delete Task By ID", func() {
		idStr := deleteIDEntry.Text

		req, err := http.NewRequest(http.MethodDelete, getBaseURL()+"/tasks/"+idStr, nil)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error creating request: %v", err))
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			updateResultLabel(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			updateResultLabel(resp.StatusCode, "Task deleted successfully!")
		} else {
			updateResultLabel(resp.StatusCode, "Failed to delete task")
		}
	})

	actionsContainer := container.NewVBox(
		widget.NewLabelWithStyle("Add New Task", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		titleEntry,
		descriptionEntry,
		dueDateEntry,
		addButton,
		widget.NewLabelWithStyle("Get All Tasks", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		getAllTasksButton,
		widget.NewLabelWithStyle("Get Task By ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		idEntry,
		getTaskByIDButton,
		widget.NewLabelWithStyle("Update Task By ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		updateIDEntry,
		updateTitleEntry,
		updateDescriptionEntry,
		updateDueDateEntry,
		updateTaskByIDButton,
		widget.NewLabelWithStyle("Delete Task By ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		deleteIDEntry,
		deleteTaskByIDButton,
	)

	actionsScroll := container.NewScroll(actionsContainer)
	actionsScroll.Hide()

	portContainer = container.NewVBox(
		widget.NewLabel("Set Server Port"),
		portEntry,
		widget.NewButton("Set Port", func() {
			serverPort = portEntry.Text
			if serverPort == "" {
				updateResultLabel(http.StatusBadRequest, "Port cannot be empty")
				return
			}

			portContainer.Hide()
			actionsScroll.Show()

			updateResultLabel(http.StatusOK, fmt.Sprintf("Port set to %s", serverPort))
		}),
	)

	resultContainer := container.NewVBox(
		widget.NewLabel("Result"),
		resultLabel,
	)
	resultScroll := container.NewScroll(resultContainer)

	content := container.NewBorder(
		portContainer, // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		container.NewHSplit(
			actionsScroll,
			resultScroll,
		),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
