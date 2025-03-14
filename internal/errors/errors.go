package errors

import "errors"

var InvalidMessageId = errors.New("unexpected message with ID %d received")
var FailedToPublish = errors.New("failed to publish message")
var FailedToParse = errors.New("failed to parse message")
var FailedToLog = errors.New("failed to log info to disk")
var UnmappedLanguage = errors.New("unmapped language")
