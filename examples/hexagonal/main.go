// Command hexagonal is a demo backend laid out as ports & adapters (hexagonal /
// clean architecture) over net/http, mounting archview at /graph.
//
//	flow per context: rest (inbound adapter) -> usecase (application core,
//	via inbound port) -> postgres / gateway (outbound adapters, via outbound ports)
package main

import (
	"log"
	"net/http"

	"github.com/eularix/archview"

	bilpg "archview-example-hexagonal/internal/billing/adapter/postgres"
	bilrest "archview-example-hexagonal/internal/billing/adapter/rest"
	biluc "archview-example-hexagonal/internal/billing/usecase"
	catgw "archview-example-hexagonal/internal/catalog/adapter/gateway"
	catpg "archview-example-hexagonal/internal/catalog/adapter/postgres"
	catrest "archview-example-hexagonal/internal/catalog/adapter/rest"
	catuc "archview-example-hexagonal/internal/catalog/usecase"
)

func main() {
	mux := http.NewServeMux()

	// catalog context: rest -> usecase -> postgres + gateway (notifier)
	catalogH := catrest.NewCatalogHandler(
		catuc.NewItemService(catpg.NewItemRepository(), catgw.NewItemNotifier()),
	)
	mux.HandleFunc("GET /catalog/items", catalogH.ListItems)
	mux.HandleFunc("POST /catalog/items", catalogH.CreateItem)

	// billing context: rest -> usecase -> postgres
	billingH := bilrest.NewBillingHandler(
		biluc.NewInvoiceService(bilpg.NewInvoiceRepository()),
	)
	mux.HandleFunc("GET /billing/invoices", billingH.ListInvoices)
	mux.HandleFunc("POST /billing/invoices", billingH.CreateInvoice)

	av, err := archview.New(archview.Options{
		Root:     ".",
		BasePath: "/graph",
		Editor:   "vscode",
	})
	if err != nil {
		log.Fatal(err)
	}
	av.Mount(mux)

	log.Println("listening on :8090 — open http://localhost:8090/graph")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		log.Fatal(err)
	}
}
