package main

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const (
	NameLength = 42

	BuilderGamePersonType = iota
	BlacksmithGamePersonType
	WarriorGamePersonType
)

type Option func(*GamePerson)

type GamePerson struct {
	x, y, z int32            // 12
	gold    uint32           // 4
	attrs   uint32           // 4
	props   uint16           // 2
	name    [NameLength]byte // 42
}

const (
	manaShift       = 0
	manaMask        = 0x3FF // 10 бит на 1000 значений
	healthShift     = 10    // на 1000 значений
	healthMask      = 0x3FF
	hasHouseShift   = 20 // 1 бит
	hasGunShift     = 21 // 1 бит
	hasFamilyShift  = 22
	personTypeShift = 23 // до 8 типов персонажей, у нас 3
	personTypeMask  = 0x3
)

const (
	respectShift    = 0
	respectMask     = 0xF // 4 бита на 10 значений
	strengthShift   = 4
	strengthMask    = 0xF
	experienceShift = 8
	experienceMask  = 0xF
	levelShift      = 12
	levelMask       = 0xF
)

func makeMask[T uint32 | uint16](mask, shift T) T {
	return mask << shift
}

func setAttrField(person *GamePerson, value uint32, mask, shift uint32) {
	person.attrs = (person.attrs & ^makeMask(mask, shift)) |
		(value&mask)<<shift
}

func setPropField(person *GamePerson, value uint16, mask, shift uint16) {
	person.props = (person.props & ^makeMask(mask, shift)) |
		(value&mask)<<shift
}

func WithName(name string) Option {
	return func(person *GamePerson) {
		copy(person.name[:], name)
	}
}

func WithCoordinates(x, y, z int) Option {
	return func(person *GamePerson) {
		person.x = int32(x)
		person.y = int32(y)
		person.z = int32(z)
	}
}

func WithGold(gold int) Option {
	return func(person *GamePerson) {
		person.gold = uint32(gold)
	}
}

func WithType(personType int) Option {
	return func(person *GamePerson) {
		setAttrField(person, uint32(personType), personTypeMask, personTypeShift)
	}
}

func WithMana(mana int) Option {
	return func(person *GamePerson) {
		setAttrField(person, uint32(mana), manaMask, manaShift)
	}
}

func WithHealth(health int) Option {
	return func(person *GamePerson) {
		setAttrField(person, uint32(health), healthMask, healthShift)
	}
}

func WithRespect(respect int) Option {
	return func(person *GamePerson) {
		setPropField(person, uint16(respect), respectMask, respectShift)
	}
}

func WithStrength(strength int) Option {
	return func(person *GamePerson) {
		setPropField(person, uint16(strength), strengthMask, strengthShift)
	}
}

func WithExperience(experience int) Option {
	return func(person *GamePerson) {
		setPropField(person, uint16(experience), experienceMask, experienceShift)
	}
}

func WithLevel(level int) Option {
	return func(person *GamePerson) {
		setPropField(person, uint16(level), levelMask, levelShift)
	}
}

func WithHouse() Option {
	return func(person *GamePerson) {
		person.attrs |= 1 << hasHouseShift
	}
}

func WithGun() Option {
	return func(person *GamePerson) {
		person.attrs |= 1 << hasGunShift
	}
}

func WithFamily() Option {
	return func(person *GamePerson) {
		person.attrs |= 1 << hasFamilyShift
	}
}

func (p *GamePerson) Name() string {
	return string(p.name[:])
}

func (p *GamePerson) X() int {
	return int(p.x)
}

func (p *GamePerson) Y() int {
	return int(p.y)
}

func (p *GamePerson) Z() int {
	return int(p.z)
}

func (p *GamePerson) Gold() int {
	return int(p.gold)
}

func (p *GamePerson) Mana() int {
	return int((p.attrs >> manaShift) & manaMask)
}

func (p *GamePerson) Health() int {
	return int((p.attrs >> healthShift) & healthMask)
}

func (p *GamePerson) Respect() int {
	return int((p.props >> respectShift) & respectMask)
}

func (p *GamePerson) Strength() int {
	return int((p.props >> strengthShift) & strengthMask)
}

func (p *GamePerson) Experience() int {
	return int((p.props >> experienceShift) & experienceMask)
}

func (p *GamePerson) Level() int {
	return int((p.props >> levelShift) & levelMask)
}

func (p *GamePerson) HasHouse() bool {
	return (p.attrs>>hasHouseShift)&1 != 0
}

func (p *GamePerson) HasGun() bool {
	return (p.attrs>>hasGunShift)&1 != 0
}

func (p *GamePerson) HasFamilty() bool {
	return (p.attrs>>hasFamilyShift)&1 != 0
}

func (p *GamePerson) Type() int {
	return int((p.attrs >> personTypeShift) & personTypeMask)
}

func NewGamePerson(options ...Option) GamePerson {
	p := GamePerson{}
	for _, option := range options {
		option(&p)
	}
	return p
}

type GamePersonJSON struct {
	X          int32  `json:"x"`
	Y          int32  `json:"y"`
	Z          int32  `json:"z"`
	Gold       uint32 `json:"gold"`
	Name       string `json:"name"`
	Mana       int    `json:"mana"`
	Health     int    `json:"health"`
	HasHouse   bool   `json:"has_house"`
	HasGun     bool   `json:"has_gun"`
	HasFamily  bool   `json:"has_family"`
	Type       string `json:"type"`
	Respect    int    `json:"respect"`
	Strength   int    `json:"strength"`
	Experience int    `json:"experience"`
	Level      int    `json:"level"`
}

func (p GamePerson) MarshalJSON() ([]byte, error) {
	jsonData := GamePersonJSON{
		X:          p.x,
		Y:          p.y,
		Z:          p.z,
		Gold:       p.gold,
		Name:       string(p.name[:]),
		Mana:       p.Mana(),
		Health:     p.Health(),
		HasHouse:   p.HasHouse(),
		HasGun:     p.HasGun(),
		HasFamily:  p.HasFamilty(),
		Type:       p.typeToString(),
		Respect:    p.Respect(),
		Strength:   p.Strength(),
		Experience: p.Experience(),
		Level:      p.Level(),
	}
	return json.Marshal(jsonData)
}

func (p *GamePerson) UnmarshalJSON(data []byte) error {
	personJSON := &GamePersonJSON{}
	if err := json.Unmarshal(data, personJSON); err != nil {
		return err
	}

	p.x = personJSON.X
	p.y = personJSON.Y
	p.z = personJSON.Z
	p.gold = personJSON.Gold
	copy(p.name[:], personJSON.Name)

	WithMana(personJSON.Mana)(p)
	WithHealth(personJSON.Health)(p)
	if personJSON.HasHouse {
		WithHouse()(p)
	}
	if personJSON.HasGun {
		WithGun()(p)
	}
	if personJSON.HasFamily {
		WithFamily()(p)
	}

	typeInt, err := p.stringToType(personJSON.Type)
	if err != nil {
		return err
	}
	WithType(typeInt)(p)

	WithRespect(personJSON.Respect)(p)
	WithStrength(personJSON.Strength)(p)
	WithExperience(personJSON.Experience)(p)
	WithLevel(personJSON.Level)(p)

	return nil
}

func (p *GamePerson) typeToString() string {
	switch p.Type() {
	case BuilderGamePersonType:
		return "builder"
	case BlacksmithGamePersonType:
		return "blacksmith"
	case WarriorGamePersonType:
		return "warrior"
	default:
		return "unknown"
	}
}

func (p *GamePerson) stringToType(s string) (int, error) {
	switch s {
	case "builder":
		return BuilderGamePersonType, nil
	case "blacksmith":
		return BlacksmithGamePersonType, nil
	case "warrior":
		return WarriorGamePersonType, nil
	default:
		return 0, errors.New("invalid person type")
	}
}

func TestGamePerson(t *testing.T) {
	assert.LessOrEqual(t, unsafe.Sizeof(GamePerson{}), uintptr(64))

	const x, y, z = math.MinInt32, math.MaxInt32, 0
	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const personType = BuilderGamePersonType
	const gold = math.MaxInt32
	const mana = 1000
	const health = 1000
	const respect = 10
	const strength = 10
	const experience = 10
	const level = 10

	options := []Option{
		WithName(name),
		WithCoordinates(x, y, z),
		WithGold(gold),
		WithMana(mana),
		WithHealth(health),
		WithRespect(respect),
		WithStrength(strength),
		WithExperience(experience),
		WithLevel(level),
		WithHouse(),
		WithFamily(),
		WithType(personType),
	}

	person := NewGamePerson(options...)
	jsonBytes, err := person.MarshalJSON()
	assert.NoError(t, err)

	personFromJSON := NewGamePerson()
	err = personFromJSON.UnmarshalJSON(jsonBytes)
	assert.NoError(t, err)

	assert.Equal(t, name, personFromJSON.Name())
	assert.Equal(t, x, personFromJSON.X())
	assert.Equal(t, y, personFromJSON.Y())
	assert.Equal(t, z, personFromJSON.Z())
	assert.Equal(t, gold, personFromJSON.Gold())
	assert.Equal(t, mana, personFromJSON.Mana())
	assert.Equal(t, health, personFromJSON.Health())
	assert.Equal(t, respect, personFromJSON.Respect())
	assert.Equal(t, strength, personFromJSON.Strength())
	assert.Equal(t, experience, personFromJSON.Experience())
	assert.Equal(t, level, personFromJSON.Level())
	assert.True(t, personFromJSON.HasHouse())
	assert.True(t, personFromJSON.HasFamilty())
	assert.False(t, personFromJSON.HasGun())
	assert.Equal(t, personType, personFromJSON.Type())
}
