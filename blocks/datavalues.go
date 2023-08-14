package blocks

type BlockData byte

// Valid on Sapling, Log, and Leaves
const (
	Oak BlockData = iota
	Spruce
	Birch
)

// Valid on FlowingWater, Water, FlowingLava, and Lava
const (
	FluidFull BlockData = iota
	FluidDrained1
	FluidDrained2
	FluidDrained3
	FluidDrained4
	FluidDrained5
	FluidDrained6
	FluidDrained7
)

// Valid only on Dispenser
const (
	DispenserNorth BlockData = iota + 2
	DispenserSouth
	DispenserWest
	DispenserEast
)

// Valid only on Bed. If less than `BedUpper`, the lower half of the bed will
// be shown.
const (
	BedNorth BlockData = iota
	BedEast
	BedSouth
	BedWest
	// Can be mixed with the above constants to make the bed show the pillow
	BedUpper BlockData = 8
)

// Valid on PoweredRail, DetectorRail, and Rail, EXCEPT for diagonal
// directions, which is only valid for Rail. PoweredRail can also use
// `RailIsPowered` in addition to the directional modifiers
const (
	RailNS BlockData = iota
	RailWE
	RailUpEast
	RailUpWest
	RailUpNorth
	RailUpSouth
	RailSE
	RailSW
	RailNW
	RailNE
)

// Valid only for PoweredRail
const RailIsPowered = 8

// Valid for StickyPiston, Piston, and PistonHead. See also `PistonExtended`.
// See also `StickyPistonHead`
const (
	PistonDown BlockData = iota
	PistonUp
	PistonNorth
	PistonSouth
	PistonWest
	PistonEast
)

// Valid for StickyPiston and Piston
const PistonExtended = 8

// Valid only for TallGrass
const (
	DeadTallGrass BlockData = iota
	GrassShrub
	Fern
)

// Valid only for PistonHead
const StickyPistonHead = 8

// Valid only for Wool
const (
	WhiteWool BlockData = iota
	OrangeWool
	MagentaWool
	LightBlueWool
	YellowWool
	LimeWool
	PinkWool
	GrayWool
	LightGrayWool
	CyanWool
	PurpleWool
	BlueWool
	BrownWool
	GreenWool
	RedWool
	BlackWool
)

// Valid for DoubleSlab and Slab
const (
	StoneSlab BlockData = iota
	SandstoneSlab
	WoodSlab
	CobblestoneSlab
)

// Valid for Torch, RedstoneTorchOff, and RedstoneTorchOn
const (
	TorchFloor BlockData = iota
	TorchEast
	TorchWest
	TorchSouth
	TorchNorth
)

// Valid for WoodStairs and CobblestoneStairs
const (
	StairsEast BlockData = iota
	StairsWest
	StairsSouth
	StairsNorth
)

// Valid only for Redstone
const (
	RedstoneOff BlockData = iota
	Redstone1
	Redstone2
	Redstone3
	Redstone4
	Redstone5
	Redstone6
	Redstone7
	Redstone8
	Redstone9
	Redstone10
	Redstone11
	Redstone12
	Redstone13
	Redstone14
	Redstone15
)

// Valid only for Wheat
const (
	Wheat1 BlockData = iota
	Wheat2
	Wheat3
	Wheat4
	Wheat5
	Wheat6
	Wheat7
	Wheat8
)

// Valid only for Farmland
const (
	DryFarmland BlockData = iota
	WetFarmland
)

// Valid for Furnace and LitFurnace
const (
	FurnaceNorth BlockData = iota + 2
	FurnaceSouth
	FurnaceWest
	FurnaceEast
)

// Valid only for StandingSign
const (
	StandingSignS BlockData = iota
	StandingSignSSW
	StandingSignSW
	StandingSignWSW
	StandingSignW
	StandingSignWNW
	StandingSignNW
	StandingSignNNW
	StandingSignN
	StandingSignNNE
	StandingSignNE
	StandingSignENE
	StandingSignE
	StandingSignESE
	StandingSignSE
	StandingSignSSE
)

// Valid for WoodenDoor and IronDoor. Door defaults to closed, bottom texture.
// Use `DoorOpen` and `DoorTop` for other cases.
const (
	DoorNW BlockData = iota
	DoorNE
	DoorSE
	DoorSW
	DoorOpen BlockData = 4
	DoorTop  BlockData = 8
)

// Valid only for Ladder
const (
	LadderNorth BlockData = iota + 2
	LadderSouth
	LadderWest
	LadderEast
)

// Valid only for WallSign
const (
	WallSignNorth BlockData = iota + 2
	WallSignSouth
	WallSignWest
	WallSignEast
)

// Valid only for Lever. Direction indicates which way the block is facing, not
// from where it is mounted.
const (
	LeverEast BlockData = iota + 1
	LeverWest
	LeverSouth
	LeverNorth
	LeverFloorNS
	LeverFloorWE
	LeverOn BlockData = 8
)

// Valid for StonePressurePlate and WoodPressurePlate
const (
	PressurePlateOff BlockData = iota
	PressurePlateOn  BlockData = iota
)

// Valid only for Lever. Direction indicates which way the block is facing, not
// from where it is mounted.
const (
	ButtonEast BlockData = iota + 1
	ButtonWest
	ButtonSouth
	ButtonNorth
	BottonPowered BlockData = 8
)

// Valid only for SnowLayer
const (
	Snow1 BlockData = iota
	Snow2
	Snow3
	Snow4
	Snow5
	Snow6
	Snow7
	Snow8
)

// Valid for Pumpkin and JackOLantern
const (
	PumpkinSouth BlockData = iota
	PumpkinWest
	PumpkinNorth
	PumpkinEast
)

// Valid only for Cake
const (
	CakeFull BlockData = iota
	Cake6
	Cake5
	Cake4
	Cake3
	Cake2
	Cake1
)

// Valid for RepeaterOff and RepeaterOn. Add `RepeaterLevel` * (0..3) for
// incrementing the repeater.
const (
	RepeaterNorth BlockData = iota
	RepeaterEast
	RepeaterSouth
	RepeaterWest
	RepeaterLevel BlockData = 4
)

// Valid only for Trapdoor. Direction indicates which way the block is facing,
// not from where it is mounted.
const (
	TrapdoorNorth BlockData = iota
	TrapdoorSouth
	TrapdoorWest
	TrapdoorEast
	TrapdoorUp BlockData = 4
)
