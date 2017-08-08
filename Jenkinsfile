def label = "jenkins.${env.JOB_NAME.replace("/","_")}.${env.BUILD_NUMBER}"

podTemplate(label: label, 
    containers: [
        containerTemplate(name: 'jnlp', image: 'jenkinsci/jnlp-slave:2.62-alpine', args: '${computer.jnlpmac} ${computer.name}'),
        containerTemplate(name: 'golang', image: 'golang:1-alpine', ttyEnabled: true, command: 'cat'),
        containerTemplate(name: 'docker', image: 'docker:1.11.2', ttyEnabled: true, command: 'cat')
    ],
    volumes: [
        hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock'),
        emptyDirVolume(mountPath: '/go/src')
    ]
) {
    node(label) {
        stage('Checkout project') {
            checkout scm
            sh "git fetch --tags"
        }

        stage('Prep build containers') {
            container('golang') {
                sh "apk update && apk add bash git make"
            }
            container('docker') {
                sh "apk update && apk add bash git make"
            }
        }

        stage('Info') {
            container('golang') {
                sh "USER=jenkins GIT_BRANCH=${env.BRANCH_NAME} make info"
            }
        }

        stage('Build') {
            container('golang') {
                sh "mkdir -p /go/src/github.com/joelanford"
                sh "ln -s `pwd` /go/src/github.com/joelanford/scm-bot"
                sh "(cd /go/src/github.com/joelanford/scm-bot && USER=jenkins make build)"
            }
        }

        stage ('Test') {
            container('golang') {
                sh "(cd /go/src/github.com/joelanford/scm-bot && make test)"
            }
        }

        stage ('Build Docker Image') {
            container('docker') {
                sh "make image"
            }
        }

        stage ('Push Docker Image') {
            container('docker') {
                withCredentials([usernamePassword(credentialsId: 'joelanford-dockerhub',
                                                usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                    sh "docker login -u '${env.USERNAME}' -p '${env.PASSWORD}'"
                    sh "GIT_BRANCH=${env.BRANCH_NAME} make push"
                }
            }
        }
    }
}