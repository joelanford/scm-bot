def label = "scm-bot"

podTemplate(label: label, 
    containers: [
        containerTemplate(name: 'jnlp', image: 'jenkinsci/jnlp-slave:2.62-alpine', args: '${computer.jnlpmac} ${computer.name}'),
        containerTemplate(name: 'golang', image: 'golang:1-alpine', ttyEnabled: true, command: 'cat')
        containerTemplate(name: 'docker', image: 'docker:stable-git', ttyEnabled: true, command: 'cat')
    ],
    volumes: [
        hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock'),
        emptyDirVolume(mountPath: '/go/src')
    ]
) {
    node(label) {
        stage('Checkout project') {
            checkout scm
        }

        stage('Prep build containers') {
            container('golang') {
                sh "apk update && apk add bash git make"
            }
            container('docker') {
                sh "apk update && apk add bash make"
            }
        }

        stage('Build') {
            container('golang') {
                sh "mkdir -p /go/src/github.com/joelanford"
                sh "ln -s `pwd` /go/src/github.com/joelanford/scm-bot"
                sh "make -C /go/src/github.com/joelanford/scm-bot build"
            }
        }

        stage ('Test') {
            container('golang') {
                sh "make -C /go/src/github.com/joelanford/scm-bot test"
            }
        }

        stage ('Build Docker Image') {
            container('docker') {
                withCredentials() {
                    sh "docker login -u ${env.USERNAME} -p ${env.PASSWORD}"
                    sh "make -C /go/src/github.com/joelanford/scm-bot image"
                }
            }
        }

        stage ('Push Docker Image') {
            container('docker') {
                sh "GIT_BRANCH=${env.BRANCH_NAME} make -C /go/src/github.com/joelanford/scm-bot push"
            }
        }
    }
}