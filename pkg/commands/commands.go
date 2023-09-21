package commands

import (
	"fmt"
	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	cloudstorages "github.com/st-tech/gatling-operator/pkg/cloudstorages"
)

func GetGatlingWaiterCommand(parallelism *int32, gatlingNamespace string, gatlingName string) string {
	template := `
PARALLELISM=%d
NAMESPACE=%s
JOB_NAME=%s
POD_NAME=$(cat /etc/pod-info/name)

kubectl label pods -n $NAMESPACE $POD_NAME gatling-waiter=initialized

while true; do
  READY_PODS=$(kubectl get pods -n $NAMESPACE --selector=job-name=$JOB_NAME-runner,gatling-waiter=initialized --no-headers | grep -c ".*");
  echo "$READY_PODS/$PARALLELISM pods are ready";
  if  [ $READY_PODS -eq $PARALLELISM ]; then
    break;
  fi;
  sleep 1;
done
`
	return fmt.Sprintf(template,
		*parallelism,
		gatlingNamespace,
		gatlingName,
	)
}

func GetGradleGatlingRunnerCommand(config gatlingv1alpha1.GatlingRunnerConfig) string {

	fmt.Println("SimulationsDirectoryPath:", config.SimulationsDirectoryPath)
	fmt.Println("ResultsDirectoryPath:", config.ResultsDirectoryPath)
	fmt.Println("SimulationClass:", config.SimulationClass)
	fmt.Println("ResourceFileName:", config.ResourceFileName)
	fmt.Println("Environment:", config.Environment)
	fmt.Println("StartTime:", config.StartTime)
	fmt.Println("TimeZone:", config.TimeZone)

	template := `

START_TIME="%s"
TIME_ZONE="%s"

run_simulation(){
    CURRENT_TIME=$1 
	START_TIME=$2
    
    SIMULATIONS_DIR_PATH="%s"
	RESULTS_DIR_PATH="%s"
	SIMULATION_CLASS="%s"
    RESOURCE_FILE_NAME="%s"
	ENVIRONMENT="%s"

	NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace 2>/dev/null)
	if [ -z "${NAMESPACE}" ]; then
	  NAMESPACE="default"
	fi

	if [[ "${RESOURCE_FILE_NAME}" == *.conf ]]; then
	  TARGET_API="${RESOURCE_FILE_NAME%%.conf}"
	else
	  TARGET_API="${RESOURCE_FILE_NAME}"
	fi

	echo "RESOURCE_FILE_NAME: ${RESOURCE_FILE_NAME}"
	echo "TARGET_API: ${TARGET_API}"
	
	RESULTS_DIR_PATH="${RESULTS_DIR_PATH}/${ENVIRONMENT}/${NAMESPACE}/${SIMULATION_CLASS}/${TARGET_API}"
	RUN_STATUS_FILE="${RESULTS_DIR_PATH}/PENDING"
	
	# Log all the parameters
	echo "SIMULATIONS_DIR_PATH: ${SIMULATIONS_DIR_PATH}"
	echo "TEMP_SIMULATIONS_DIR_PATH: ${TEMP_SIMULATIONS_DIR_PATH}"
	echo "RESULTS_DIR_PATH: ${RESULTS_DIR_PATH}"
	echo "RUN_STATUS_FILE: ${RUN_STATUS_FILE}"
	
	echo "I am running in namespace $NAMESPACE"
	
    echo "Wait until ${START_TIME}"
	until [[ "$CURRENT_TIME" > "$START_TIME" ]]; do
	  CURRENT_TIME=$(date "+%%Y-%%m-%%d %%H:%%M:%%S")
	  echo "it's ${CURRENT_TIME} now and waiting until ${START_TIME} ..."
	  sleep 1
	done
	
	if [ ! -d ${SIMULATIONS_DIR_PATH} ]; then
	  mkdir -p ${SIMULATIONS_DIR_PATH}
	fi

	
	if [ ! -d ${RESULTS_DIR_PATH} ]; then
	  mkdir -p ${RESULTS_DIR_PATH}
	fi

	echo "Running Gatling simulation with gradle..."
    echo "cd ${SIMULATIONS_DIR_PATH} && gradle gatlingRun-${SIMULATION_CLASS} -PconfigResource=${RESOURCE_FILE_NAME} && gradle moveGatlingReports -PresultsFolder=${RESULTS_DIR_PATH}"
	cd ${SIMULATIONS_DIR_PATH} && gradle gatlingRun-${SIMULATION_CLASS} -PconfigResource=${RESOURCE_FILE_NAME} && gradle moveGatlingReports zipGatlingReports -PresultsFolder=${RESULTS_DIR_PATH}
	sleep 60
	
	GRADLE_EXIT_STATUS=$?
	if [ $GRADLE_EXIT_STATUS -ne 0 ]; then
	  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/FAILED"
	  echo "Gradle tasks have failed!" 1>&2
	else
	  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/COMPLETED"
	  echo "Gradle tasks completed successfully."
	fi
	touch ${RUN_STATUS_FILE}
	exit $GRADLE_EXIT_STATUS
}


echo "Running Gatling simulation with  gradle"

if [ -z "${TIME_ZONE}" ]; then
  TIME_ZONE="UTC"
fi

if [ -f "/usr/share/zoneinfo/${TIME_ZONE}" ]; then
  export TZ="${TIME_ZONE}"
  echo "Using ${TIME_ZONE} timezone"
  sleep 1
else
  echo "Time zone ${TIME_ZONE} is not valid. Falling back to UTC."
  export TZ="UTC"
  sleep 1
fi

CHECK_DATE_IN_PAST="true"
if [ -z "${START_TIME}" ]; then
  START_TIME=$(date +"%%Y-%%m-%%d %%H:%%M:%%S")
CHECK_DATE_IN_PAST="false"
fi

CURRENT_TIME=$(date "+%%Y-%%m-%%d %%H:%%M:%%S")


echo "Start Date: ${START_TIME}"
echo "Current Date: ${CURRENT_TIME}"

# Check if START_TIME is in the past
if [[ "$CHECK_DATE_IN_PAST" == "true" &&  "$CURRENT_TIME" > "$START_TIME" ]] ;then 
  echo "Start time is in the past, nothing is going to be executed."
else
  echo "Checking the start date. It is in the future or now, proceeding with the script."
  run_simulation "$CURRENT_TIME" "$START_TIME"
fi
`
	return fmt.Sprintf(template,
		config.StartTime,
		config.TimeZone,
		config.SimulationsDirectoryPath,
		config.ResultsDirectoryPath,
		config.SimulationClass,
		config.ResourceFileName,
		config.Environment,
	)
}

func GetGatlingRunnerCommand(
	simulationsDirectoryPath string, tempSimulationsDirectoryPath string, resourcesDirectoryPath string,
	resultsDirectoryPath string, startTime string, timezone string, simulationClass string, generateLocalReport bool) string {

	template := `
SIMULATIONS_DIR_PATH=%s
TEMP_SIMULATIONS_DIR_PATH=%s
RESOURCES_DIR_PATH=%s
RESULTS_DIR_PATH=%s
START_TIME="%s"
RUN_STATUS_FILE="${RESULTS_DIR_PATH}/COMPLETED"
if [ -z "${START_TIME}" ]; then
  START_TIME=$(date +"%%Y-%%m-%%d %%H:%%M:%%S" --utc)
fi
start_time_stamp=$(date -d "${START_TIME}" +"%%s")
current_time_stamp=$(date +"%%s")
echo "Wait until ${START_TIME}"
until [ ${current_time_stamp} -ge ${start_time_stamp} ];
do
  current_time_stamp=$(date +"%%s")
  echo "it's ${current_time_stamp} now and waiting until ${start_time_stamp} ..."
  sleep 1;
done
if [ ! -d ${SIMULATIONS_DIR_PATH} ]; then
  mkdir -p ${SIMULATIONS_DIR_PATH}
fi
if [ -d ${TEMP_SIMULATIONS_DIR_PATH} ]; then
  cp -p ${TEMP_SIMULATIONS_DIR_PATH}/*.scala ${SIMULATIONS_DIR_PATH}
fi
if [ ! -d ${RESOURCES_DIR_PATH} ]; then
  mkdir -p ${RESOURCES_DIR_PATH}
fi
if [ ! -d ${RESULTS_DIR_PATH} ]; then
  mkdir -p ${RESULTS_DIR_PATH}
fi
gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s %s -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH} %s

GATLING_EXIT_STATUS=$?
if [ $GATLING_EXIT_STATUS -ne 0 ]; then
  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/FAILED"
  echo "gatling.sh has failed!" 1>&2
fi
touch ${RUN_STATUS_FILE}
exit $GATLING_EXIT_STATUS
`
	generateLocalReportOption := "-nr"
	if generateLocalReport {
		generateLocalReportOption = ""
	}

	return fmt.Sprintf(template,
		simulationsDirectoryPath,
		tempSimulationsDirectoryPath,
		resourcesDirectoryPath,
		resultsDirectoryPath,
		startTime,
		simulationClass,
		generateLocalReportOption)
}

func GetGatlingTransferResultCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingTransferAllResultCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}

func GetGatlingAggregateResultCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingAggregateResultCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}

func GetGatlingGenerateReportCommand(resultsDirectoryPath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
DIR_NAME=$(dirname ${GATLING_AGGREGATE_DIR})
BASE_NAME=$(basename ${GATLING_AGGREGATE_DIR})
gatling.sh -rf ${DIR_NAME} -ro ${BASE_NAME}
`
	return fmt.Sprintf(template, resultsDirectoryPath)
}

func GetGatlingTransferReportCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingTransferReportCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}
