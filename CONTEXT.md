# Gold Shop

This context owns how a Thai gold shop prices, sells, buys back, pawns and safekeeps gold for its customers across its branches. It is the glossary for the Gold-shop product (`gold-shop-api`, `gold-shop-web`); identity terms (User, Client, System, Role) come from UM and are not redefined here.

## Language

### Gold and its price

**Gold Price**:
The shop's own price of the day, in four numbers: what it pays and what it charges for bar gold, and the same pair for ornament gold. Each Client sets its own, by hand or from the market feed; it is never shared between shops.
_Avoid_: Market price, spot, rate, quote

**Price Snapshot**:
The exact Gold Price a Sale or Pawn was made at, copied onto the transaction so the record never changes when the price does.
_Avoid_: Current price (on a transaction), historical price

**Kind**:
Whether gold is bar or ornament. Kind decides which pair of the Gold Price applies, how weight converts between baht-weight and grams, and whether a Category exists.
_Avoid_: Type (ambiguous with Gold Type), form

**Gold Type**:
The purity grade of the gold, such as 96.5% or 99.99%.
_Avoid_: Karat, fineness, Kind

**Category**:
The shape of an ornament: necklace, bracelet, ring, bangle, earring, pendant or amulet casing. Bars have no Category.
_Avoid_: Product type, style, Design (the free-text pattern of a piece)

**Baht-weight**:
The Thai unit gold is priced in. One baht-weight of bar gold and one of ornament gold are different numbers of grams, so the conversion depends on Kind.
_Avoid_: Baht (the currency), weight

### Stock

**Product**:
A catalog entry for a design of gold the shop sells: Kind, Gold Type, Category, name and pictures. A Product holds no quantity and no price of its own; its price follows the Gold Price.
_Avoid_: Item, SKU (its code), listing

**Item**:
One physical piece of a Product, identified by barcode and serial number, with its own weight and status: available, reserved, sold, or pawned. Stock is counted in Items.
_Avoid_: Unit, piece, stock, Product

**Stock Movement**:
One change to where an Item is or whether it is for sale: added, removed, sold, cancelled back, transferred, or adjusted, always with who did it and which transaction caused it.
_Avoid_: Stock log, inventory event, audit

**Transfer**:
Moving Items from one Branch to another, requested at one end and received at the other, and in transit in between.
_Avoid_: Shipment, relocation

### Selling and buying

**Sale**:
One transaction at the counter at a Price Snapshot: selling Items, buying old gold from the customer, or exchanging the two in one go, settled by one or more Payments. Points are earned and spent on a Sale.
_Avoid_: Order, invoice, transaction, bill

**Sale Type**:
Which way gold moves in a Sale: sell (shop to customer), buy old (customer to shop) or exchange (both).
_Avoid_: Mode, direction

**Old Gold**:
Gold a customer brings in to sell or trade, described by Kind, Gold Type, weight and condition (good, fair or damaged). Condition decides the deduction before the shop pays.
_Avoid_: Scrap (one Destination), trade-in, used gold

**Destination**:
Where Old Gold goes after a Sale: melted, put back into stock to resell, or kept as scrap.
_Avoid_: Disposal, routing

**Buyback**:
The price the shop pays for Old Gold: the Gold Price's buy side, by weight, less the deduction for condition.
_Avoid_: Purchase price, offer

**Payment**:
One way the customer settled part of a Sale: cash, credit card, bank transfer or voucher, with its amount and reference.
_Avoid_: Tender, settlement

### Pawn and saving

**Pawn**:
A loan secured by the customer's gold: the Items pledged, the amount lent, the monthly interest rate and a term. It is active, extended, redeemed or forfeited.
_Avoid_: Loan, collateral, mortgage

**Interest Period**:
The 30-day block interest is counted in. Interest accrues only for whole matured periods; a partial period owes nothing yet.
_Avoid_: Month (calendar month), cycle

**Redemption**:
The customer paying the Pawn off, principal plus interest due less any discount, and taking the gold back.
_Avoid_: Payoff, closing, settlement

**Forfeit**:
The shop taking ownership of the pledged gold after a Pawn is not redeemed in time.
_Avoid_: Default, seizure

**Gold Saving**:
A customer's account at the shop for accumulating gold over time by deposits and withdrawals, each in cash or in physical gold, until it is closed.
_Avoid_: Savings plan, installment, wallet

### Customers and the shop

**Customer**:
A person the shop sells to, buys from, lends to or saves for, identified by phone, ID card or member card. A Customer may be a Member.
_Avoid_: Client (the UM word), buyer, account

**Member**:
A Customer who has joined the loyalty program: a member code, a tier (bronze, silver, gold, platinum) and a points balance.
_Avoid_: VIP, loyalty customer

**Reward**:
Something a Member may claim with points while it is valid and in stock.
_Avoid_: Gift, redemption (the Pawn word), coupon

**Branch**:
One shop location of a Client with its own Items, Sales, Pawns, Expenses and Employees. The head office is the Branch an unassigned Employee falls back to.
_Avoid_: Store, outlet, tenant

**Employee**:
A UM User working at a Branch with a shop role: admin, manager or staff. The shop role is granted here and is not the UM Role.
_Avoid_: User (the UM word), staff (one of the roles), worker

**Expense**:
Money the Branch spent, in a category, recorded for approval: pending, approved or rejected.
_Avoid_: Cost, outgoing, bill
