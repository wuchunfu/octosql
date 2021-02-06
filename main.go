package main

import (
	"context"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/cube2222/octosql/datasources/json"
	"github.com/cube2222/octosql/execution"
	"github.com/cube2222/octosql/execution/aggregates"
	"github.com/cube2222/octosql/parser"
	"github.com/cube2222/octosql/parser/sqlparser"
	"github.com/cube2222/octosql/physical"
)

func main() {
	statement, err := sqlparser.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	logicalPlan, err := parser.ParseSelect(statement.(*sqlparser.Select))
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Wrap panics into errors in subfunction.
	physicalPlan := logicalPlan.Typecheck(
		context.Background(),
		physical.Environment{
			Aggregates: map[string][]physical.AggregateDescriptor{
				"count": aggregates.CountOverloads,
			},
			Datasources: &physical.DatasourceRepository{
				Datasources: map[string]func(name string) (physical.DatasourceImplementation, error){
					"json": json.Creator,
				},
			},
			PhysicalConfig:  nil,
			VariableContext: nil,
		},
		physical.State{},
	)
	spew.Dump(physicalPlan.Schema)
	executionPlan, err := physicalPlan.Materialize(
		context.Background(),
		physical.Environment{
			Aggregates: nil,
			Datasources: &physical.DatasourceRepository{
				Datasources: map[string]func(name string) (physical.DatasourceImplementation, error){
					"json": json.Creator,
				},
			},
			PhysicalConfig:  nil,
			VariableContext: nil,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := executionPlan.Run(
		execution.ExecutionContext{
			Context:         context.Background(),
			VariableContext: nil,
		},
		func(ctx execution.ProduceContext, record execution.Record) error {
			log.Println(record.String())
			return nil
		},
		func(ctx execution.ProduceContext, msg execution.MetadataMessage) error {
			log.Printf("%+v", msg)
			return nil
		},
	); err != nil {
		log.Fatal(err)
	}
}
