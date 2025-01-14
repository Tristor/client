// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package systests

import (
	"testing"

	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	"github.com/keybase/client/go/service"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	context "golang.org/x/net/context"
)

type trackingUI struct {
	signupUI
}

func (n *trackingUI) GetIdentifyTrackUI() libkb.IdentifyUI {
	return &identifyUI{}
}

type identifyUI struct {
}

func (*identifyUI) Confirm(*keybase1.IdentifyOutcome) (keybase1.ConfirmResult, error) {
	return keybase1.ConfirmResult{
		IdentityConfirmed: true,
		RemoteConfirmed:   true,
	}, nil
}
func (*identifyUI) Start(string, keybase1.IdentifyReason)                                 {}
func (*identifyUI) FinishWebProofCheck(keybase1.RemoteProof, keybase1.LinkCheckResult)    {}
func (*identifyUI) FinishSocialProofCheck(keybase1.RemoteProof, keybase1.LinkCheckResult) {}
func (*identifyUI) DisplayCryptocurrency(keybase1.Cryptocurrency)                         {}
func (*identifyUI) DisplayKey(keybase1.IdentifyKey)                                       {}
func (*identifyUI) ReportLastTrack(*keybase1.TrackSummary)                                {}
func (*identifyUI) LaunchNetworkChecks(*keybase1.Identity, *keybase1.User)                {}
func (*identifyUI) DisplayTrackStatement(string) error                                    { return nil }
func (*identifyUI) DisplayUserCard(keybase1.UserCard)                                     {}
func (*identifyUI) ReportTrackToken(keybase1.TrackToken) error                            { return nil }
func (*identifyUI) SetStrict(b bool)                                                      {}
func (*identifyUI) Finish()                                                               {}

func (*identifyUI) DisplayTLFCreateWithInvite(keybase1.DisplayTLFCreateWithInviteArg) error {
	return nil
}

type trackingNotifyHandler struct {
	trackingCh chan keybase1.TrackingChangedArg
	errCh      chan error
}

func newTrackingNotifyHandler() *trackingNotifyHandler {
	return &trackingNotifyHandler{
		trackingCh: make(chan keybase1.TrackingChangedArg),
		errCh:      make(chan error),
	}
}

func (h *trackingNotifyHandler) TrackingChanged(_ context.Context, arg keybase1.TrackingChangedArg) error {
	h.trackingCh <- arg
	return nil
}

func TestTrackingNotifications(t *testing.T) {
	tc := setupTest(t, "signup")
	tc2 := cloneContext(tc)
	tc5 := cloneContext(tc)

	libkb.G.LocalDb = nil

	// Hack the various portions of the service that aren't
	// properly contextified.

	defer tc.Cleanup()

	stopCh := make(chan error)
	svc := service.NewService(tc.G, false)
	startCh := svc.GetStartChannel()
	go func() {
		err := svc.Run()
		if err != nil {
			t.Logf("Running the service produced an error: %v", err)
		}
		stopCh <- err
	}()

	userInfo := randomUser("sgnup")

	tui := trackingUI{
		signupUI: signupUI{
			info:         userInfo,
			Contextified: libkb.NewContextified(tc2.G),
		},
	}
	tc2.G.SetUI(&tui)
	signup := client.NewCmdSignupRunner(tc2.G)
	signup.SetTest()

	<-startCh

	if err := signup.Run(); err != nil {
		t.Fatal(err)
	}
	tc2.G.Log.Debug("Login State: %v", tc2.G.LoginState())

	nh := newTrackingNotifyHandler()

	// Launch the server that will listen for tracking notifications.
	launchServer := func(nh *trackingNotifyHandler) error {
		cli, xp, err := client.GetRPCClientWithContext(tc5.G)
		if err != nil {
			return err
		}
		srv := rpc.NewServer(xp, nil)
		if err = srv.Register(keybase1.NotifyTrackingProtocol(nh)); err != nil {
			return err
		}
		ncli := keybase1.NotifyCtlClient{Cli: cli}
		if err = ncli.SetNotifications(context.TODO(), keybase1.NotificationChannels{
			Tracking: true,
		}); err != nil {
			return err
		}
		return nil
	}

	// Actually launch it in the background
	go func() {
		err := launchServer(nh)
		if err != nil {
			nh.errCh <- err
		}
	}()

	// Have our test user track t_alice.
	trackCmd := client.NewCmdTrackRunner(tc2.G)
	trackCmd.SetUser("t_alice")
	trackCmd.SetOptions(keybase1.TrackOptions{BypassConfirm: true})
	err := trackCmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Do a check for new tracking statements that should fire off a
	// notification. Currently the track command above does not fetch the new
	// chain link from the server, so this call is required. It's possible that
	// TrackEngine (or our signature caching code) might change in the future,
	// making this call unnecessary.
	checkTrackingCmd := client.NewCmdCheckTrackingRunner(tc2.G)
	err = checkTrackingCmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Wait to get a notification back as we expect.
	// NOTE: If this test ever starts deadlocking here, it's possible that
	// we've changed how we cache signatures that we make on the local client,
	// in such a way that the fetch done by CheckTracking above doesn't find
	// any "isOwnNewLinkFromServer" links. If so, one way to fix this test
	// would be to blow away the local db before calling CheckTracking.
	tc.G.Log.Debug("Waiting for tracking notification.")
	select {
	case err := <-nh.errCh:
		t.Fatalf("Error before notify: %v", err)
	case arg := <-nh.trackingCh:
		tAliceUID := keybase1.UID("295a7eea607af32040647123732bc819")
		tc.G.Log.Debug("Got tracking changed notification (%#v)", arg)
		if "t_alice" != arg.Username {
			t.Fatalf("Bad username back: %s != %s", "t_alice", arg.Username)
		}
		if !tAliceUID.Equal(arg.Uid) {
			t.Fatalf("Bad UID back: %s != %s", tAliceUID, arg.Uid)
		}
	}

	stopper := client.NewCmdCtlStopRunner(tc2.G)
	if err := stopper.Run(); err != nil {
		t.Fatal(err)
	}

	// If the server failed, it's also an error
	if err := <-stopCh; err != nil {
		t.Fatal(err)
	}
}
