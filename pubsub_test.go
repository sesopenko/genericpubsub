package genericpubsub

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

const testBufferSize = 64

func TestPubSub_RegisterShouldNotifySub(t *testing.T) {
	type message struct {
		Value string
	}

	// Arrange
	input := message{
		Value: "hello",
	}
	expected := message{
		Value: "hello",
	}
	ps := New[message](context.TODO(), testBufferSize)
	events := ps.Register(context.TODO(), testBufferSize)

	// Act
	go func() {
		ps.Send(input)
	}()

	// Assert
	wasClosed := true
	select {
	case result := <-events:
		wasClosed = false
		assert.Equal(t, expected.Value, result.Value)
	case <-time.After(200 * time.Millisecond):
		wasClosed = false
		t.Fatalf("Timed out waiting for event")
	}
	assert.False(t, wasClosed, "Should not have closed the channel")
}

func TestPubSub_RegisterShouldNotifyLateSubscriber(t *testing.T) {
	// Given 2 subs
	// And 2nd sub has a delayed subscription
	// And 2nd sub is slow to read channel
	// Should not delay 1st sub
	// And 2nd sub should only read 2nd message
	type message struct {
		Value string
	}

	// Arrange
	firstInput := message{
		Value: "first",
	}
	secondInput := message{
		Value: "second",
	}
	firstExpected := message{
		Value: "first",
	}
	secondExpected := message{
		Value: "second",
	}
	ps := New[message](context.TODO(), testBufferSize)
	firstEvents := ps.Register(context.TODO(), testBufferSize)

	// Send first message
	ps.Send(firstInput)
	time.Sleep(50 * time.Millisecond)

	select {
	case result, ok := <-firstEvents:
		assert.True(t, ok)
		assert.Equal(t, firstExpected, result)
	}

	time.Sleep(100 * time.Millisecond)
	secondEvents := ps.Register(context.TODO(), testBufferSize)

	secondDone := make(chan bool)
	go func() {
		// Make the 2nd sub sluggish, should not block other sub
		time.Sleep(50 * time.Millisecond)
		select {
		case secondResultSeconSub, ok := <-secondEvents:
			assert.True(t, ok)
			assert.Equal(t, secondExpected, secondResultSeconSub)
		case <-time.After(200 * time.Millisecond):
			// This is a pass scenario, should receive no more messages
		}
		secondDone <- true
	}()

	sendTime := time.Now()
	ps.Send(secondInput)

	select {
	case result, ok := <-firstEvents:
		elapsed := time.Since(sendTime)
		assert.LessOrEqual(t, elapsed, 10*time.Millisecond, "Should not be held up by other subscribers")
		assert.True(t, ok)
		assert.Equal(t, secondExpected, result)
	case <-time.After(500 * time.Millisecond):
		t.Errorf("Timed out waiting for 2nd event for 1st sub")
	}

	select {
	case <-secondDone:
		log.Printf("second done")
		select {
		case _, ok := <-secondEvents:
			assert.False(t, ok, "Should not have received a 2nd event on 2nd sub")
		case <-time.After(400 * time.Millisecond):
			// do nothing, this is expected
		}
	case <-time.After(400 * time.Millisecond):
		// do nothing, this is expected
	}

}

func TestPubSub_RegisterShouldNotNotifyCancelledSub(t *testing.T) {
	type message struct {
		Value string
	}
	input := message{
		Value: "hello",
	}
	cancellableCtx, cancel := context.WithCancel(context.Background())
	ps := New[message](context.TODO(), testBufferSize)
	cancel()
	events := ps.Register(cancellableCtx, testBufferSize)
	go func() {
		time.Sleep(100 * time.Millisecond)
		ps.Send(input)
	}()
	select {
	case _, chanWasOpen := <-events:
		if chanWasOpen {
			t.Fatalf("Should not receive a message after cancellation of context.")
		}
		break
	case <-time.After(200 * time.Millisecond):
		// We've given it enough time for the message send and should
		// now reach this point.  Simply break, because we pass if this
		// happens.
		break
	}
}

func TestPubSub_RegisterShouldNotNotifyCancelledSystem(t *testing.T) {
	type message struct {
		Value string
	}
	input := message{
		Value: "hello",
	}
	ctx, cancel := context.WithCancel(context.Background())
	ps := New[message](ctx, testBufferSize)
	cancel()
	events := ps.Register(context.TODO(), testBufferSize)
	go func() {
		time.Sleep(100 * time.Millisecond)
		ps.Send(input)
	}()
	select {
	case _, chanWasOpen := <-events:
		if chanWasOpen {
			t.Fatalf("Should not receive a message after cancellation of context.")
		}
		break
	case <-time.After(200 * time.Millisecond):
		// We've given it enough time for the message send and should
		// now reach this point.  Simply break, because we pass if this
		// happens.
		break
	}
}
