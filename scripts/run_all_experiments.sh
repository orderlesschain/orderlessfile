#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env
start=$(date +%s)

rm -rf "${PROJECT_ABSOLUTE_PATH}"/configs/certs
mkdir "${PROJECT_ABSOLUTE_PATH}"/configs/certs
if [[ $BUILD_MODE == "local" ]]; then
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_local.yml "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints.yml
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/certs_local/* "${PROJECT_ABSOLUTE_PATH}"/configs/certs
else
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_remote.yml "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints.yml
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/certs_remote/* "${PROJECT_ABSOLUTE_PATH}"/configs/certs
fi

commitResults(){
    pushd "${PROJECT_ABSOLUTE_PATH}"/orderlessfile-experiments
    git add .
    git commit -m 'auto push'
    git pull origin master
    git push origin master
    popd
}


runFileThroughputOne(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in file.scale1 file.scale2 file.scale3 file.scale4 file.scale5 file.scale6 file.scale7 file.scale8 ; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runFileThroughputTwo(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in file.scale9 file.scale10 file.scale11 file.scale12 file.scale13 file.scale14 file.scale15 file.scale16 ; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runFileThroughputThree(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in file.scale17 file.scale18 file.scale19 file.scale20 file.scale21 file.scale22 file.scale23 file.scale24 ; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runFileThroughputFour(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in file.scale25 file.scale26 file.scale27 file.scale28 file.scale29 file.scale30 file.scale31 file.scale32 ; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runAllFile(){
  runFileThroughputOne

  runFileThroughputTwo

  runFileThroughputThree

  runFileThroughputFour
}

runAllFile

runAllFile

runAllFile


end=$(date +%s)
echo Experiments executed in $(expr $end - $start) seconds.
