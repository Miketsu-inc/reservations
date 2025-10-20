export const unitOptions = [
  { value: "g", label: "g" },
  { value: "dkg", label: "dkg" },
  { value: "kg", label: "kg" },
  { value: "ml", label: "ml" },
  { value: "dl", label: "dl" },
  { value: "L", label: "L" },
  { value: "pcs", label: "pcs" },
];

export const unitConversionMap = {
  g: { base: "g", factor: 1, type: "mass" },
  dkg: { base: "g", factor: 10, type: "mass" },
  kg: { base: "g", factor: 1000, type: "mass" },

  ml: { base: "ml", factor: 1, type: "volume" },
  dl: { base: "ml", factor: 100, type: "volume" },
  L: { base: "ml", factor: 1000, type: "volume" },

  pcs: { base: "pcs", factor: 1, type: "count" },
};

export const convertToBaseUnit = (value, unit) => {
  const { factor } = unitConversionMap[unit];
  return parseInt(value) * factor;
};

export const convertFromBaseUnit = (baseValue, baseUnit) => {
  const unitInfo = unitConversionMap[baseUnit];
  if (!unitInfo) return { value: baseValue, unit: baseUnit };

  const type = unitInfo.type;

  const units = Object.entries(unitConversionMap)
    .filter(([_, info]) => info.type === type)
    .sort((a, b) => b[1].factor - a[1].factor);

  for (const [unit, { factor }] of units) {
    const converted = baseValue / factor;

    if (Number.isInteger(converted)) {
      return { value: converted, unit };
    }
  }

  return { value: baseValue, unit: baseUnit };
};

export function getDisplayUnit(current, max, baseUnit) {
  const unitType = unitConversionMap[baseUnit].type;

  const units = Object.entries(unitConversionMap)
    .filter(([_, val]) => val.type === unitType)
    .sort((a, b) => b[1].factor - a[1].factor);

  for (const [unit, { factor }] of units) {
    const currentConverted = current / factor;
    const maxConverted = max / factor;

    if (currentConverted >= 1 && maxConverted >= 1) {
      return {
        current: parseFloat(currentConverted.toFixed(2)),
        max: parseFloat(maxConverted.toFixed(2)),
        unit,
      };
    }
  }
  return { current, max, unit: baseUnit };
}
