package feedback

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"benchy/internal/domain/ports"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// ConsoleFeedback implémente l'interface FeedbackService pour la console
type ConsoleFeedback struct {
	level ports.FeedbackLevel
	colors bool
}

// NewConsoleFeedback crée un nouveau service de feedback console
func NewConsoleFeedback() *ConsoleFeedback {
	return &ConsoleFeedback{
		level:  ports.FeedbackLevelNormal,
		colors: true,
	}
}

// Info affiche un message d'information
func (cf *ConsoleFeedback) Info(ctx context.Context, message string) error {
	fmt.Println(message)
	return nil
}

// Success affiche un message de succès
func (cf *ConsoleFeedback) Success(ctx context.Context, message string) error {
	if cf.colors {
		color.Green(message)
	} else {
		fmt.Println(message)
	}
	return nil
}

// Warning affiche un message d'avertissement
func (cf *ConsoleFeedback) Warning(ctx context.Context, message string) error {
	if cf.colors {
		color.Yellow(message)
	} else {
		fmt.Println(message)
	}
	return nil
}

// Error affiche un message d'erreur
func (cf *ConsoleFeedback) Error(ctx context.Context, message string) error {
	if cf.colors {
		color.Red(message)
	} else {
		fmt.Fprintln(os.Stderr, message)
	}
	return nil
}

// StartProgress démarre un tracker de progression
func (cf *ConsoleFeedback) StartProgress(ctx context.Context, title string, total int) (ports.ProgressTracker, error) {
	return &ConsoleProgressTracker{title: title, total: total}, nil
}

// StartSpinner démarre un spinner
func (cf *ConsoleFeedback) StartSpinner(ctx context.Context, message string) (ports.Spinner, error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()
	return &ConsoleSpinner{spinner: s}, nil
}

// DisplayTable affiche un tableau
func (cf *ConsoleFeedback) DisplayTable(ctx context.Context, headers []string, rows [][]string) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	return nil
}

// DisplayJSON affiche des données JSON
func (cf *ConsoleFeedback) DisplayJSON(ctx context.Context, data interface{}) error {
	fmt.Printf("%+v\n", data)
	return nil
}

// Confirm demande une confirmation
func (cf *ConsoleFeedback) Confirm(ctx context.Context, message string) (bool, error) {
	fmt.Printf("%s (y/N): ", message)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y", nil
}

// Input demande une saisie
func (cf *ConsoleFeedback) Input(ctx context.Context, prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)
	var input string
	fmt.Scanln(&input)
	return input, nil
}

// Types pour Progress et Spinner
type ConsoleProgressTracker struct {
	title string
	total int
	current int
}

func (cpt *ConsoleProgressTracker) Update(current int, message string) error {
	fmt.Printf("📊 %s: %d/%d - %s\n", cpt.title, current, cpt.total, message)
	return nil
}

func (cpt *ConsoleProgressTracker) Increment(message string) error {
	cpt.current++
	return cpt.Update(cpt.current, message)
}

func (cpt *ConsoleProgressTracker) Complete(message string) error {
	fmt.Printf("✅ %s: Completed - %s\n", cpt.title, message)
	return nil
}

func (cpt *ConsoleProgressTracker) Error(message string) error {
	fmt.Printf("❌ %s: Failed - %s\n", cpt.title, message)
	return nil
}

func (cpt *ConsoleProgressTracker) Close() error {
	return nil
}

type ConsoleSpinner struct {
	spinner *spinner.Spinner
}

func (cs *ConsoleSpinner) UpdateMessage(message string) error {
	cs.spinner.Suffix = " " + message
	return nil
}

func (cs *ConsoleSpinner) Success(message string) error {
	cs.spinner.Stop()
	color.Green("✅ " + message)
	return nil
}

func (cs *ConsoleSpinner) Error(message string) error {
	cs.spinner.Stop()
	color.Red("❌ " + message)
	return nil
}

func (cs *ConsoleSpinner) Stop() error {
	cs.spinner.Stop()
	return nil
}
