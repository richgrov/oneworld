package blocks

type Block struct {
	Type BlockType
	Data byte
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
	DoubleStoneSlab
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
	Door
	Ladder
	Rail
	StoneStairs
	WallSign
	Lever
	StonePressurePlate
	IronDoor
	WoodPressurePlat
	RedstoneOre
	PoweredRedstoneO
	RedstoneTorchOff
	RedstoneTorchOn
	StoneButton
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
