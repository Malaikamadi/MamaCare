import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  Image,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Alert,
} from 'react-native';
import { Stack, router } from 'expo-router';
import { Colors } from '@/constants/Colors';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { StatusBar } from 'expo-status-bar';

const SIERRA_LEONE_COUNTRY_CODE = '+232';

export default function LoginScreen(): React.JSX.Element {
  const [phoneNumber, setPhoneNumber] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  
  // Function to validate Sierra Leone phone number format
  const isValidPhoneNumber = (phone: string): boolean => {
    // Sierra Leone numbers are typically +232 XX XXX XXX
    const phoneRegex = /^[7-9]\d{7}$/; // 8 digits starting with 7, 8, or 9
    return phoneRegex.test(phone);
  };
  
  const handleLogin = (): void => {
    // Remove any spaces or dashes
    const cleanedNumber = phoneNumber.replace(/[-\s]/g, '');
    
    if (!isValidPhoneNumber(cleanedNumber)) {
      Alert.alert(
        'Invalid Phone Number',
        'Please enter a valid Sierra Leone phone number, e.g. 76123456',
        [{ text: 'OK' }]
      );
      return;
    }
    
    // Show loading state
    setIsLoading(true);
    
    // Simulate authentication process
    setTimeout(() => {
      setIsLoading(false);
      
      // Navigate to the main app
      router.replace('/(tabs)');
    }, 1500);
  };
  
  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      keyboardVerticalOffset={Platform.OS === 'ios' ? 0 : 20}
    >
      <StatusBar style="dark" />
      <Stack.Screen options={{ headerShown: false }} />
      
      <ScrollView 
        contentContainerStyle={styles.scrollContent}
        keyboardShouldPersistTaps="handled"
      >
        <View style={styles.header}>
          {/* Sierra Leone flag colors - Green, White, Blue */}
          <View style={styles.flagContainer}>
            <View style={[styles.flagStripe, { backgroundColor: '#1EB53A' }]} />
            <View style={[styles.flagStripe, { backgroundColor: '#FFFFFF' }]} />
            <View style={[styles.flagStripe, { backgroundColor: '#0072C6' }]} />
          </View>
          <Text style={styles.title}>MamaCare</Text>
          <Text style={styles.subtitle}>Sierra Leone Maternal Health</Text>
        </View>
        
        <View style={styles.formContainer}>
          <Text style={styles.welcomeText}>Welcome, Mother!</Text>
          <Text style={styles.instructionText}>
            Enter your phone number to access your pregnancy journey
          </Text>
          
          <View style={styles.phoneInputContainer}>
            <View style={styles.countryCodeContainer}>
              <Text style={styles.countryCodeText}>{SIERRA_LEONE_COUNTRY_CODE}</Text>
            </View>
            <TextInput
              style={styles.phoneInput}
              placeholder="76 123 456"
              keyboardType="phone-pad"
              value={phoneNumber}
              onChangeText={setPhoneNumber}
              maxLength={9} // Country code is separate
              autoFocus
            />
          </View>
          
          <TouchableOpacity 
            style={[
              styles.loginButton,
              isLoading && styles.loginButtonDisabled
            ]}
            onPress={handleLogin}
            disabled={isLoading}
          >
            {isLoading ? (
              <Text style={styles.loginButtonText}>Signing In...</Text>
            ) : (
              <Text style={styles.loginButtonText}>Sign In</Text>
            )}
          </TouchableOpacity>
          
          <TouchableOpacity style={styles.helpContainer}>
            <IconSymbol name="phone.fill" size={16} color="#777" />
            <Text style={styles.helpText}>Need help? Call 099 Support</Text>
          </TouchableOpacity>
        </View>
        
        <View style={styles.footer}>
          <Text style={styles.footerText}>
            Your pregnancy journey companion
          </Text>
          <Text style={styles.versionText}>
            Version 1.0.0
          </Text>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'space-between',
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginTop: 60,
    marginBottom: 40,
  },
  flagContainer: {
    width: 120,
    height: 60,
    flexDirection: 'row',
    marginBottom: 20,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 4,
    overflow: 'hidden',
  },
  flagStripe: {
    flex: 1,
    height: '100%',
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#E91E63',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#555',
  },
  formContainer: {
    width: '100%',
    alignItems: 'center',
  },
  welcomeText: {
    fontSize: 24,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
    textAlign: 'center',
  },
  instructionText: {
    fontSize: 16,
    color: '#666',
    marginBottom: 32,
    textAlign: 'center',
    lineHeight: 22,
  },
  phoneInputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 24,
    width: '100%',
  },
  countryCodeContainer: {
    backgroundColor: '#f0f0f0',
    height: 56,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 16,
    borderTopLeftRadius: 8,
    borderBottomLeftRadius: 8,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRightWidth: 0,
  },
  countryCodeText: {
    fontSize: 16,
    color: '#333',
    fontWeight: '500',
  },
  phoneInput: {
    flex: 1,
    height: 56,
    backgroundColor: '#f9f9f9',
    paddingHorizontal: 16,
    fontSize: 16,
    color: '#333',
    borderWidth: 1,
    borderColor: '#ddd',
    borderTopRightRadius: 8,
    borderBottomRightRadius: 8,
  },
  loginButton: {
    backgroundColor: '#E91E63',
    width: '100%',
    height: 56,
    borderRadius: 28,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 24,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  loginButtonDisabled: {
    backgroundColor: '#f0a0b5',
  },
  loginButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  helpContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  helpText: {
    color: '#777',
    marginLeft: 8,
    fontSize: 14,
  },
  footer: {
    alignItems: 'center',
    marginTop: 40,
    marginBottom: 20,
  },
  footerText: {
    color: '#888',
    fontSize: 14,
    marginBottom: 8,
  },
  versionText: {
    color: '#ccc',
    fontSize: 12,
  },
});
