package alliance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	simulation2 "github.com/terra-money/alliance/x/alliance/tests/simulation"

	// this line is used by starport scaffolding # 1

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/terra-money/alliance/x/alliance/client/cli"
	"github.com/terra-money/alliance/x/alliance/keeper"
	migrationsv4 "github.com/terra-money/alliance/x/alliance/migrations/v4"
	"github.com/terra-money/alliance/x/alliance/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModule           = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.EndBlockAppModule   = AppModule{}
)

type AppModuleBasic struct {
	cdc  codec.Codec
	pcdc *codec.ProtoCodec
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, sk types.StakingKeeper,
	ak types.AccountKeeper, bk types.BankKeeper, registry cdctypes.InterfaceRegistry,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc, pcdc: codec.NewProtoCodec(registry)},
		keeper:         keeper,
		stakingKeeper:  sk,
		bankKeeper:     bk,
		accountKeeper:  ak,
	}
}

func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (a AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (a AppModuleBasic) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(DefaultGenesisState())
}

func (a AppModuleBasic) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	var genesis types.GenesisState
	if err := jsonCodec.UnmarshalJSON(message, &genesis); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return ValidateGenesis(&genesis)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)) //nolint:errcheck
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for the alliance module.
type AppModule struct {
	AppModuleBasic
	keeper        keeper.Keeper
	stakingKeeper types.StakingKeeper
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
}

func (a AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, a.keeper)
}

func (a AppModule) InitGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	jsonCodec.MustUnmarshalJSON(message, &genesis)
	return a.keeper.InitGenesis(ctx, &genesis)
}

func (a AppModule) ExportGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	genesis := a.keeper.ExportGenesis(ctx)
	return a.cdc.MustMarshalJSON(genesis)
}

func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {
	RegisterInvariants(registry, a.keeper)
}

// Deprecated: use RegisterServices
func (a AppModule) Route() sdk.Route { return sdk.Route{} }

// Deprecated: use RegisterServices
func (AppModule) QuerierRoute() string { return types.RouterKey }

// Deprecated: use RegisterServices
func (a AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(a.keeper))
	err := cfg.RegisterMigration(types.ModuleName, 3, migrationsv4.Migrate(a.keeper))
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/alliance from version 3 to 4: %v", err))
	}
}

func (a AppModule) ConsensusVersion() uint64 {
	return 4
}

func (a AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation2.RandomizedGenesisState(simState)
}

func (a AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

func (a AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return simulation2.ParamChanges(r)
}

func (a AppModule) RegisterStoreDecoder(registry sdk.StoreDecoderRegistry) {
	registry[types.StoreKey] = simulation2.NewDecodeStore(a.cdc)
}

func (a AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation2.WeightedOperations(simState.AppParams, a.pcdc, a.accountKeeper, a.bankKeeper, a.stakingKeeper, a.keeper)
}
