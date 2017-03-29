package main

import (
	"strconv"

	"github.com/cyverse-de/dockerops"
	"github.com/cyverse-de/logcabin"
	"github.com/cyverse-de/messaging"
	"github.com/cyverse-de/model"
)

// RemoveVolume removes the docker volume with the provided volume identifier.
func RemoveVolume(id string) {
	var (
		err       error
		hasVolume bool
	)
	hasVolume, err = dckr.VolumeExists(id)
	if err != nil {
		logcabin.Error.Print(err)
	}
	if hasVolume {
		logcabin.Info.Printf("removing volume: %s", id)
		if err = dckr.RemoveVolume(id); err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// RemoveJobContainers removes containers based on their job label.
func RemoveJobContainers(id string) {
	logcabin.Info.Printf("Finding all containers with the label %s=%s", model.DockerLabelKey, id)
	jobContainers, err := dckr.ContainersWithLabel(model.DockerLabelKey, id, true)
	if err != nil {
		logcabin.Error.Print(err)
		jobContainers = []string{}
	}
	for _, jc := range jobContainers {
		logcabin.Info.Printf("Nuking container %s", jc)
		err = dckr.NukeContainer(jc)
		if err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// RemoveDataContainers attempts to remove all data containers.
func RemoveDataContainers() {
	logcabin.Info.Println("Finding all data containers")
	dataContainers, err := dckr.ContainersWithLabel(dockerops.TypeLabel, strconv.Itoa(dockerops.DataContainer), true)
	if err != nil {
		logcabin.Error.Print(err)
	}
	for _, dc := range dataContainers {
		logcabin.Info.Printf("Nuking data container %s", dc)
		err = dckr.NukeContainer(dc)
		if err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// RemoveStepContainers attempts to remove all step containers.
func RemoveStepContainers() {
	logcabin.Info.Println("Finding all step containers")
	stepContainers, err := dckr.ContainersWithLabel(dockerops.TypeLabel, strconv.Itoa(dockerops.StepContainer), true)
	if err != nil {
		logcabin.Error.Print(err)
	}
	for _, sc := range stepContainers {
		logcabin.Info.Printf("Nuking step container %s", sc)
		err = dckr.NukeContainer(sc)
		if err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// RemoveInputContainers attempts to remove all input containers.
func RemoveInputContainers() {
	logcabin.Info.Println("Finding all input containers")
	inputContainers, err := dckr.ContainersWithLabel(dockerops.TypeLabel, strconv.Itoa(dockerops.InputContainer), true)
	if err != nil {
		logcabin.Error.Print(err)
		inputContainers = []string{}
	}
	for _, ic := range inputContainers {
		logcabin.Info.Printf("Nuking input container %s", ic)
		err = dckr.NukeContainer(ic)
		if err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// RemoveDataContainerImages removes the images for the data containers.
func RemoveDataContainerImages() {
	var err error
	for _, dc := range job.DataContainers() {
		logcabin.Info.Printf("Nuking image %s:%s", dc.Name, dc.Tag)
		err = dckr.NukeImage(dc.Name, dc.Tag)
		if err != nil {
			logcabin.Error.Print(err)
		}
	}
}

// cleanup encapsulates common job clean up tasks.
func cleanup(job *model.Job) {
	logcabin.Info.Printf("Performing aggressive clean up routine...")
	RemoveInputContainers()
	RemoveStepContainers()
	RemoveDataContainers()
	RemoveVolume(job.InvocationID)
}

// Exit handles clean up when road-runner is killed.
func Exit(exit, finalExit chan messaging.StatusCode) {
	exitCode := <-exit
	logcabin.Warning.Printf("Received an exit code of %d, cleaning up", int(exitCode))
	switch exitCode {
	case messaging.StatusKilled:
		//Annihilate the input/steps/data containers even if they're running,
		//but allow the output containers to run. Yanking the rug out from the
		//containers should force the Run() function to 'fall through' to any clean
		//up steps.
		RemoveDataContainerImages()
		RemoveInputContainers()
		RemoveStepContainers()
		RemoveDataContainers()
		RemoveVolume(job.InvocationID)
		RemoveJobContainers(job.InvocationID)

	default:
		RemoveJobContainers(job.InvocationID)
		RemoveVolume(job.InvocationID)
	}

	finalExit <- exitCode
}
