// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package framework

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/allinbits/apcore/app"
	"github.com/allinbits/apcore/framework/oauth2"
	"github.com/allinbits/apcore/framework/web"
	"github.com/allinbits/apcore/paths"
	"github.com/allinbits/apcore/services"
	"github.com/allinbits/apcore/util"
	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
)

var _ app.Framework = &Framework{}

type Framework struct {
	scheme            string
	host              string
	rsaKeySize        int
	saltSize          int
	bCryptStrength    int
	o                 *oauth2.Server
	s                 *web.Sessions
	data              *services.Data
	followers         *services.Followers
	users             *services.Users
	actor             pub.Actor
	federationEnabled bool
}

func BuildFramework(scheme string,
	host string,
	rsaKeySize int,
	saltSize int,
	bCryptStrength int,
	fw *Framework,
	o *oauth2.Server,
	s *web.Sessions,
	data *services.Data,
	followers *services.Followers,
	users *services.Users,
	actor pub.Actor,
	a app.Application) *Framework {
	_, isS2S := a.(app.S2SApplication)
	fw.scheme = scheme
	fw.host = host
	fw.rsaKeySize = rsaKeySize
	fw.saltSize = saltSize
	fw.bCryptStrength = bCryptStrength
	fw.o = o
	fw.s = s
	fw.data = data
	fw.actor = actor
	fw.federationEnabled = isS2S
	fw.followers = followers
	fw.users = users
	return fw
}

func (f *Framework) Context(r *http.Request) context.Context {
	return util.WithAPHTTPContext(f.scheme, f.host, r)
}

func (f *Framework) CreateUser(c context.Context, username, email, password string) (userID string, err error) {
	p := services.CreateUserParameters{
		Scheme:     f.scheme,
		Host:       f.host,
		RSAKeySize: f.rsaKeySize,
		HashParams: services.HashPasswordParameters{
			SaltSize:       f.saltSize,
			BCryptStrength: f.bCryptStrength,
		},
		Username: username,
		Email:    email,
	}
	ctx := util.Context{c}
	return f.users.CreateUser(ctx, p, password)
}

func (f *Framework) IsNotUniqueUsername(err error) bool {
	return err == services.NotUniqueUsername
}

func (f *Framework) IsNotUniqueEmail(err error) bool {
	return err == services.NotUniqueEmail
}

func (f *Framework) UserIRI(userUUID paths.UUID) *url.URL {
	return paths.UUIDIRIFor(f.scheme, f.host, paths.UserPathKey, userUUID)
}

func (f *Framework) Validate(w http.ResponseWriter, r *http.Request) (userID paths.UUID, authenticated bool, err error) {
	var suID string
	suID, authenticated, err = f.o.Validate(w, r)
	userID = paths.UUID(suID)
	return
}

func (f *Framework) Send(c context.Context, userID paths.UUID, t vocab.Type) error {
	ctx := util.Context{c}
	ctx.WithUserPathUUID(userID)
	if !f.federationEnabled {
		return fmt.Errorf("cannot Send: Framework.Send called when federation is not enabled")
	} else if fa, ok := f.actor.(pub.FederatingActor); !ok {
		return fmt.Errorf("cannot Send: pub.Actor is not a pub.FederatingActor with federation enabled")
	} else {
		outboxIRI := paths.UUIDIRIFor(f.scheme, f.host, paths.OutboxPathKey, userID)
		_, err := fa.Send(ctx.Context, outboxIRI, t)
		return err
	}
}

func (f *Framework) GetPrivileges(c context.Context, userID paths.UUID, appPrivileges interface{}) (admin bool, err error) {
	var p *services.Privileges
	p, err = f.users.Privileges(util.Context{c}, string(userID), appPrivileges)
	if err != nil {
		return
	}
	admin = p.Admin
	return
}

func (f *Framework) SetPrivileges(c context.Context, userID paths.UUID, admin bool, appPrivileges interface{}) error {
	p := &services.Privileges{
		Admin:         admin,
		InstanceActor: false,
		AppPrivileges: appPrivileges,
	}
	return f.users.UpdatePrivileges(util.Context{c}, string(userID), p)
}

func (f *Framework) Session(r *http.Request) (app.Session, error) {
	return f.s.Get(r)
}

func (f *Framework) GetByIRI(c context.Context, id *url.URL) (vocab.Type, error) {
	return f.data.Get(util.Context{c}, id)
}

func (f *Framework) OpenFollowRequests(c context.Context, userID paths.UUID) ([]vocab.ActivityStreamsFollow, error) {
	return f.followers.OpenFollowRequests(util.Context{c}, f.UserIRI(userID))
}

func (f *Framework) SendAcceptFollow(ctx context.Context, userID paths.UUID, followIRI *url.URL) error {
	myIRI := f.UserIRI(userID)

	follow, err := f.getValidFollow(ctx, myIRI, followIRI)
	if err != nil {
		return err
	}

	// Build the Accept
	accept := streams.NewActivityStreamsAccept()

	me := streams.NewActivityStreamsActorProperty()
	me.AppendIRI(myIRI)
	accept.SetActivityStreamsActor(me)

	op := streams.NewActivityStreamsObjectProperty()
	op.AppendIRI(followIRI)
	accept.SetActivityStreamsObject(op)

	to := streams.NewActivityStreamsToProperty()
	followActors := follow.GetActivityStreamsActor()
	for iter := followActors.Begin(); iter != followActors.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err != nil {
			return err
		}
		to.AppendIRI(id)
	}
	accept.SetActivityStreamsTo(to)
	// Deliver the Accept
	if err := f.Send(ctx, paths.UUID(userID), accept); err != nil {
		return err
	}
	// Update the followers collection
	followersIRI := paths.UserIRIFor(f.scheme, f.host, paths.FollowersPathKey, paths.Actor(userID))
	for iter := followActors.Begin(); iter != followActors.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err != nil {
			return err
		}
		err = f.followers.PrependItem(util.Context{ctx}, followersIRI, id)
		if err != nil {
			// TODO: Soft fail instead?
			return fmt.Errorf("accepted Follow but not all actors were added to followers collection(%s)[%s]: %w", followersIRI, id, err)
		}
	}
	return nil
}

func (f *Framework) SendRejectFollow(ctx context.Context, userID paths.UUID, followIRI *url.URL) error {
	myIRI := f.UserIRI(userID)

	follow, err := f.getValidFollow(ctx, myIRI, followIRI)
	if err != nil {
		return err
	}

	// Build the Reject
	reject := streams.NewActivityStreamsReject()

	me := streams.NewActivityStreamsActorProperty()
	me.AppendIRI(myIRI)
	reject.SetActivityStreamsActor(me)

	op := streams.NewActivityStreamsObjectProperty()
	op.AppendIRI(followIRI)
	reject.SetActivityStreamsObject(op)

	to := streams.NewActivityStreamsToProperty()
	followActors := follow.GetActivityStreamsActor()
	for iter := followActors.Begin(); iter != followActors.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err != nil {
			return err
		}
		to.AppendIRI(id)
	}
	reject.SetActivityStreamsTo(to)
	// Deliver the Reject
	if err := f.Send(ctx, paths.UUID(userID), reject); err != nil {
		return err
	}
	return nil
}

func (f *Framework) getValidFollow(ctx context.Context, userIRI *url.URL, followIRI *url.URL) (vocab.ActivityStreamsFollow, error) {
	// Fetch the Follow from our database
	tFollow, err := f.GetByIRI(ctx, followIRI)
	if err != nil {
		return nil, err
	}
	follow, err := util.ToActivityStreamsFollow(util.Context{ctx}, tFollow)
	if err != nil {
		return nil, err
	}

	// Ensure myIRI is in the object of the original follow
	present := false
	obj := follow.GetActivityStreamsObject()
	if obj != nil {
		for iter := obj.Begin(); iter != obj.End(); iter = iter.Next() {
			id, err := pub.ToId(iter)
			if err != nil {
				return nil, err
			}
			if id.String() == userIRI.String() {
				present = true
				break
			}
		}
	}
	if !present {
		return nil, fmt.Errorf("cannot Accept Follow: Follow is not for this user")
	}

	return follow, nil
}
