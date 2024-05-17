package model

type LMSSyncResult struct {
	UserSync       *UserSyncResult       `json:"user-sync"`
	AssignmentSync *AssignmentSyncResult `json:"assignment-sync"`
}

type AssignmentSyncResult struct {
	SyncedAssignments     []AssignmentInfo `json:"synced-assignments"`
	AmbiguousMatches      []AssignmentInfo `json:"ambiguous-matches"`
	NonMatchedAssignments []AssignmentInfo `json:"non-matched-assignments"`
	UnchangedAssignments  []AssignmentInfo `json:"unchanged-assignments"`
}

type AssignmentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAssignmentSyncResult() *AssignmentSyncResult {
	return &AssignmentSyncResult{
		SyncedAssignments:     make([]AssignmentInfo, 0),
		AmbiguousMatches:      make([]AssignmentInfo, 0),
		NonMatchedAssignments: make([]AssignmentInfo, 0),
		UnchangedAssignments:  make([]AssignmentInfo, 0),
	}
}

type UserSyncResult struct {
	Add []*User
	Mod []*User
	Del []*User

	// Users that exist and will not be overwritten.
	Skip []*User

	// Users that could have been modified, but would not be changed.
	Unchanged []*User

	ClearTextPasswords map[string]string
}

type UserResolveResult struct {
	Add       *User
	Mod       *User
	Del       *User
	Skip      *User
	Unchanged *User

	ClearTextPassword string
}

func NewUserSyncResult() *UserSyncResult {
	return &UserSyncResult{
		Add:                make([]*User, 0),
		Mod:                make([]*User, 0),
		Del:                make([]*User, 0),
		Skip:               make([]*User, 0),
		Unchanged:          make([]*User, 0),
		ClearTextPasswords: make(map[string]string),
	}
}

func (this *UserSyncResult) Count() int {
	return len(this.Add) + len(this.Mod) + len(this.Del)
}

func (this *UserSyncResult) AddResolveResult(resolveResult *UserResolveResult) {
	if resolveResult == nil {
		return
	}

	if resolveResult.Add != nil {
		this.Add = append(this.Add, resolveResult.Add)

		if resolveResult.ClearTextPassword != "" {
			this.ClearTextPasswords[resolveResult.Add.Email] = resolveResult.ClearTextPassword
		}
	}

	if resolveResult.Mod != nil {
		this.Mod = append(this.Mod, resolveResult.Mod)
	}

	if resolveResult.Del != nil {
		this.Del = append(this.Del, resolveResult.Del)
	}

	if resolveResult.Skip != nil {
		this.Skip = append(this.Skip, resolveResult.Skip)
	}

	if resolveResult.Unchanged != nil {
		this.Unchanged = append(this.Unchanged, resolveResult.Unchanged)
	}
}

func (this *UserSyncResult) UpdateUsers(newUsers map[string]*User) {
	updateUsersList(this.Add, newUsers)
	updateUsersList(this.Mod, newUsers)
	updateUsersList(this.Del, newUsers)
	updateUsersList(this.Skip, newUsers)
	updateUsersList(this.Unchanged, newUsers)
}

func updateUsersList(oldUsers []*User, newUsers map[string]*User) {
	for i, oldUser := range oldUsers {
		newUser, ok := newUsers[oldUser.Email]
		if !ok {
			continue
		}

		oldUsers[i] = newUser
	}
}
