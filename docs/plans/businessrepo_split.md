BusinessRepo split plan

Motivation
- `BusinessRepo` interface in `internal/usecase/interfaces/interface.go` is large and mixes multiple responsibilities (domain, invites, membership, domains). Splitting reduces test coupling and eases implementation changes.

Approach
1. Identify logical aggregates: Business lifecycle (Create/Get/Update/Delete), Membership (AddUser/RemoveUser/GetUsers/GetUserRole/HasMembership), Invites (CreateInvite/GetInvite/Accept/Reject), Domains (CreateDomain/GetDomain/Verify/AutoJoin).
2. Define smaller interfaces in `internal/usecase/interfaces/` e.g. `BusinessLifecycleRepo`, `MembershipRepo`, `InviteRepo`, `DomainRepo`.
3. Update concrete repository implementation to implement the new interfaces.
4. Update usecases to depend only on the interfaces they need.
5. Add unit tests ensuring the implementation satisfies each interface.

Rollout
- Make changes in small commits per interface to minimize review friction.

