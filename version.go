package main

// appVersion — версия Pantheon. До 1.0 живём в 0.x: API/конфиг/поведение
// могут меняться без обратной совместимости.
//
// Это var, а не const, чтобы релизный тул мог подставить версию через
// -ldflags "-X main.appVersion=0.2.0".
var appVersion = "0.2.0"
