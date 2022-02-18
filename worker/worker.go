package worker

import (
	"os"
	"time"
	"sync"
	"strconv"
	"net/http"
	"io/ioutil"
	"os/signal"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enroller/dataaccess"
	"github.com/Olling/Enroller/dataaccess/config"
)

func Initialize() {
	//Start Interupt Channels
	interruptChannel, interruptWaitGroup := initializeInterupt()

	//Start Workers
    	for w := 1; w <= config.Configuration.ThreadNumber; w++ {
    	    	go worker(interruptChannel, interruptWaitGroup, w)
    	}
}

func worker(interruptChannel <-chan os.Signal, interruptWaitGroup *sync.WaitGroup, workerID int) {
	interrupt := false
	go func(){
		for sig := range interruptChannel {
			interrupt = true
			slog.PrintInfo("Worker Thread", workerID, "Received an interrupt (", sig,") - Killing Process")
			
			interruptWaitGroup.Done()
		}
	}()

	for {
		if interrupt {break}

		response, err := http.Get(config.Configuration.EnrolldURL + "/job/" + config.Configuration.WorkerID + "-thread" + strconv.Itoa(workerID))
		if err != nil { 
			slog.PrintDebug("Could not get job from:", config.Configuration.EnrolldURL + "/job", "due to error:", err)
			time.Sleep(config.Configuration.Interval * time.Second)
			continue
		}

		if response.StatusCode == 204 {
			slog.PrintTrace("There are no jobs in the queue at:", config.Configuration.EnrolldURL + "/job")
			time.Sleep(config.Configuration.Interval * time.Second)
			continue
		}
		if response.StatusCode == 401 {
			slog.PrintError("Authenication error:", config.Configuration.EnrolldURL + "/job")
			time.Sleep(config.Configuration.Interval * time.Second * 10)
			continue
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			slog.PrintError("Something went wrong trying to get a job", err)
			time.Sleep(config.Configuration.Interval * time.Second)
			continue
		}

		var job objects.Job
		utils.StructFromJson(body, &job)

		if job.JobID == "" {
			slog.PrintError("Something went wrong trying to get a job", err, response.StatusCode)

			time.Sleep(config.Configuration.Interval * time.Second)
			continue
		}

		err = dataaccess.EnrollServer(job.Server)

		client := &http.Client{}
		if err != nil{
			slog.PrintError("Failed to enroll server:", job.Server.ServerID)
			slog.PrintDebug(job)
			request,_ := http.NewRequest("DELETE",config.Configuration.EnrolldURL + "/job/" + job.WorkerID + "/" + job.JobID + "/failure", nil)
			response, err = client.Do(request)
		} else {
			slog.PrintInfo("Enrolled server:", job.Server.ServerID)
			slog.PrintDebug(job)
			request,_ := http.NewRequest("DELETE",config.Configuration.EnrolldURL + "/job/" + job.WorkerID + "/" + job.JobID + "/success", nil)
			response, err = client.Do(request)
		}

		if err != nil {
			slog.PrintError("Failed to report result:", err)
		}
    	}
}

func initializeInterupt() (<-chan os.Signal, *sync.WaitGroup) {
	//To catch an interupt by the user
	interruptChannel := make(chan os.Signal, 1)
	
	//A channel to tell the threads to stop what they are doing
	killthreadsChannel := make(chan os.Signal, 1)

	//A wait group to wait for all of the threads to stop
	var interruptWaitGroup sync.WaitGroup	
	interruptWaitGroup.Add(config.Configuration.ThreadNumber)

	//Catch user interrupt
	signal.Notify(interruptChannel, os.Interrupt)
	go func(){
	    for sig := range interruptChannel {
		slog.PrintInfo("Program was interrupted")

		//Tell the threads to interrupt
        	for a := 1; a <= config.Configuration.ThreadNumber; a++ {
			killthreadsChannel <-sig
        	}

		//Wait for all the groups to stop
		interruptWaitGroup.Wait()
		
		//Exit
		os.Exit(1)
	    }
	}()
	return killthreadsChannel, &interruptWaitGroup
}
