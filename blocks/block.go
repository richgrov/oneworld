package blocks

// Zero value is air
type Block struct {
	Type BlockType
	Data BlockData
}

type BlockType byte

const (
	Air BlockType = iota
	Stone
	Grass
	Dirt
	Cobblestone
	Planks
	Sapling
	Bedrock
	FlowingWater
	Water
	FlowingLava
	Lava
	Sand
	Gravel
	GoldOre
	IronOre
	CoalOre
	Log
	Leaves
	Sponge
	Glass
	LapisOre
	LapisBlock
	Dispenser
	Sandstone
	NoteBlock
	Bed
	PoweredRail
	DetectorRail
	StickyPiston
	Web
	TallGrass
	DeadBush
	Piston
	PistonHead
	Wool
	PistonExtension
	Dandelion
	Rose
	BrownMushroom
	RedMushroom
	GoldBlock
	IronBlock
	DoubleSlab
	Slab
	Bricks
	Tnt
	Bookshelf
	MossStone
	Obsidian
	Torch
	Fire
	Spawner
	WoodStairs
	Chest
	Redstone
	DiamondOre
	DiamondBlock
	CraftingTable
	Wheat
	Farmland
	Furnace
	LitFurnace
	StandingSign
	WoodenDoor
	Ladder
	Rail
	StoneStairs
	WallSign
	Lever
	StonePressurePlate
	IronDoor
	WoodPressurePlate
	RedstoneOre
	PoweredRedstoneOre
	RedstoneTorchOff
	RedstoneTorchOn
	Button
	SnowLayer
	Ice
	Snow
	Cactus
	Clay
	SugarCane
	Jukebox
	Fence
	Pumpkin
	Netherrack
	SoulSand
	Glowstone
	Portal
	JackOLantern
	Cake
	RepeaterOff
	RepeaterOn
	LockedChest
	Trapdoor
)
