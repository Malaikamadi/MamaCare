import { Stack } from 'expo-router';
import React from 'react';
import { Platform, View } from 'react-native';

import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

/**
 * Layout for MamaCare app screens
 * We're using a hamburger menu in each screen instead of the tab bar navigation
 */
export default function AppLayout() {
  const colorScheme = useColorScheme();

  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: {
          backgroundColor: '#f8f9fa',
        },
      }}
    >
      <Stack.Screen name="index" />
      <Stack.Screen name="appointments" />
      <Stack.Screen name="resources" />
      <Stack.Screen name="health-records" />
      <Stack.Screen name="chat" />
      <Stack.Screen name="profile" />
    </Stack>
  );
}
