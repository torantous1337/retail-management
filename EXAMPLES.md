# Product Property Examples

This file demonstrates the flexible property system used in the Retail Management System.

## Electrical Components

### Circuit Breaker
```json
{
  "name": "Industrial Circuit Breaker",
  "sku": "CB-220-50A",
  "base_price": 45.99,
  "properties": {
    "voltage": "220v",
    "amperage": "50A",
    "poles": 3,
    "trip_curve": "D",
    "manufacturer": "ABB",
    "warranty_years": 5
  }
}
```

### LED Bulb
```json
{
  "name": "LED Bulb 10W",
  "sku": "LED-10W-WW",
  "base_price": 5.99,
  "properties": {
    "wattage": "10W",
    "color_temperature": "warm_white",
    "lumens": 800,
    "base_type": "E27",
    "dimmable": true,
    "lifespan_hours": 25000
  }
}
```

## Future Category: Liquor

### Wine
```json
{
  "name": "Chateau Margaux 2015",
  "sku": "WINE-CM-2015",
  "base_price": 299.99,
  "properties": {
    "type": "red_wine",
    "varietal": "Bordeaux Blend",
    "vintage": 2015,
    "region": "Margaux, France",
    "alcohol_content": "13.5%",
    "bottle_size": "750ml",
    "rating": 95
  }
}
```

## Future Category: Pharmaceuticals

### Medication
```json
{
  "name": "Aspirin 100mg",
  "sku": "MED-ASP-100",
  "base_price": 8.99,
  "properties": {
    "active_ingredient": "Acetylsalicylic Acid",
    "dosage": "100mg",
    "form": "tablet",
    "quantity": 100,
    "prescription_required": false,
    "expiry_date": "2026-12-31",
    "manufacturer": "Bayer",
    "lot_number": "LOT123456"
  }
}
```

## Future Category: Automotive Parts

### Brake Pad
```json
{
  "name": "Ceramic Brake Pads",
  "sku": "AUTO-BP-CER-001",
  "base_price": 79.99,
  "properties": {
    "material": "ceramic",
    "compatible_vehicles": ["Toyota Camry 2018-2023", "Honda Accord 2017-2022"],
    "position": "front",
    "includes_hardware": true,
    "noise_rating": "low",
    "wear_indicator": true
  }
}
```

## Key Benefits

1. **No Schema Migration**: Add new product categories without altering the database schema
2. **Category-Specific Attributes**: Each category can have unique properties
3. **Flexible Querying**: Properties stored as JSON, queryable in SQLite 3.38+ using JSON functions
4. **Type Safety**: Application layer handles validation and typing
5. **Future-Proof**: Easy to extend to new business domains
