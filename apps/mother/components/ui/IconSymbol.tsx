// Custom icon component that uses MaterialIcons

import MaterialIcons from '@expo/vector-icons/MaterialIcons';
import { StyleProp, TextStyle, OpaqueColorValue } from 'react-native';

/**
 * Define our custom icon mapping to provide developer-friendly icon names.
 * This also allows for more readable icon names in our codebase.
 */
const ICON_MAPPING = {
  // Navigation & UI
  'house.fill': 'home',
  'paperplane.fill': 'send',
  'chevron.left.forwardslash.chevron.right': 'code',
  'chevron.right': 'chevron-right',
  'magnifyingglass': 'search',
  'plus': 'add',
  'square.and.arrow.up': 'share',
  'calendar': 'event',
  'bell.fill': 'notifications',
  'person.fill': 'person',
  'book.fill': 'book',
  'waveform.path.ecg': 'monitor_heart',
  'heart.text.square.fill': 'favorite',
  'heart.fill': 'favorite',
  'message.fill': 'message',
  'clock.fill': 'access_time',
  'map': 'map',
  'mappin.and.ellipse': 'location_on',
  'person.badge.shield.checkmark.fill': 'verified_user',
  'person.2.fill': 'people',
  'exclamationmark.triangle.fill': 'warning',
  'drop.fill': 'opacity',
  'bed.double.fill': 'bed',
  'leaf.fill': 'eco',
  'stethoscope': 'medical_services',
  'cross.case.fill': 'medical_services',
  'syringe': 'vaccines',
  'pill.fill': 'medication',
  'ruler.fill': 'straighten',
  'doc.text.fill': 'description',
  'waveform': 'graphic_eq',
  'phone.fill': 'phone',
  'envelope.fill': 'email',
  'scalemass.fill': 'monitor_weight',
  'location.fill': 'place',
  'building.2.fill': 'apartment',
  'figure.pregnant': 'pregnant_woman',
  'mug.fill': 'coffee',
  'figure.and.child.holdinghands': 'family_restroom',
  'play.circle.fill': 'play_circle',
  'rectangle.portrait.and.arrow.right': 'logout',
  'calendar.badge.exclamationmark': 'event_busy',
  'calendar.badge.clock': 'event_available',
} as const;

// Get the type from the mapping keys for predetermined icons
type MappedIconName = keyof typeof ICON_MAPPING;

// Allow both mapped icons and direct string values
type IconName = MappedIconName | string;

/**
 * Props for the IconSymbol component
 */
interface IconSymbolProps {
  name: IconName;
  size?: number;
  color: string | OpaqueColorValue;
  style?: StyleProp<TextStyle>;
  weight?: 'regular' | 'bold' | 'light' | 'medium' | 'semibold';
}

/**
 * Fallback to convert dynamic icon names to Material Icons format
 * This provides a sensible default for common dynamic icon names
 */
const convertToMaterialIcon = (iconName: string): string => {
  // Handle function-generated icon names by returning a sensible default
  return iconName.replace(/\./g, '_').toLowerCase();
};

/**
 * An icon component that uses a combination of predefined mappings and dynamic fallbacks.
 * This allows both type-safe usage with predefined icons and flexibility for dynamic content.
 */
export function IconSymbol({ name, size = 24, color, style, weight }: IconSymbolProps) {
  // Check if the icon name is in our mapping
  const isMappedIcon = (iconName: string): iconName is MappedIconName => {
    return iconName in ICON_MAPPING;
  };
  
  // Get the appropriate material icon name
  let materialIconName: string;
  
  if (isMappedIcon(name)) {
    // Use our predefined mapping for known icons
    materialIconName = ICON_MAPPING[name];
  } else {
    // For dynamic strings, try to convert to a sensible Material icon name
    materialIconName = convertToMaterialIcon(name);
  }
  
  // Render the icon with the determined name
  return <MaterialIcons name={materialIconName as any} size={size} color={color} style={style} />;
}
