import React from 'react';
import { View, Text, StyleSheet, SafeAreaView, ScrollView, TouchableOpacity, Switch, Alert, Image } from 'react-native';
import { Stack, router } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define interfaces for type safety
interface UserProfile {
  id: string;
  name: string;
  phoneNumber: string;
  email: string | null;
  dueDate: string;
  bloodType: string | null;
  emergencyContact: {
    name: string;
    phoneNumber: string;
    relationship: string;
  } | null;
  language: 'english' | 'krio' | 'mende' | 'temne';
  region: string;
  healthFacility: string;
  assignedChw: string;
}

interface NotificationSetting {
  id: string;
  type: string;
  enabled: boolean;
  description: string;
}

// Mock data for user profile
const USER_PROFILE: UserProfile = {
  id: '12345',
  name: 'Aminata Kamara',
  phoneNumber: '+232 76 123 456',
  email: 'aminata.k@example.com',
  dueDate: '2025-12-15',
  bloodType: 'O+',
  emergencyContact: {
    name: 'Ibrahim Kamara',
    phoneNumber: '+232 77 234 567',
    relationship: 'Husband',
  },
  language: 'english',
  region: 'Freetown, Western Area',
  healthFacility: 'Freetown Community Clinic',
  assignedChw: 'Mary Johnson',
};

// Mock data for notification settings
const NOTIFICATION_SETTINGS: NotificationSetting[] = [
  {
    id: '1',
    type: 'appointmentReminders',
    enabled: true,
    description: 'Appointment reminders (24 hrs before)',
  },
  {
    id: '2',
    type: 'medicationReminders',
    enabled: true,
    description: 'Medication reminders',
  },
  {
    id: '3',
    type: 'healthTips',
    enabled: true,
    description: 'Daily health tips and advice',
  },
  {
    id: '4',
    type: 'weeklyUpdates',
    enabled: false,
    description: 'Weekly pregnancy progress updates',
  },
  {
    id: '5',
    type: 'communityMessages',
    enabled: false,
    description: 'Community messages and announcements',
  },
];

export default function ProfileScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [notificationSettings, setNotificationSettings] = 
    React.useState<NotificationSetting[]>(NOTIFICATION_SETTINGS);

  const toggleNotification = (id: string): void => {
    setNotificationSettings(prevSettings => {
      return prevSettings.map(setting => {
        if (setting.id === id) {
          return { ...setting, enabled: !setting.enabled };
        }
        return setting;
      });
    });
  };
  
  const handleLogout = (): void => {
    Alert.alert(
      'Logout',
      'Are you sure you want to log out?',
      [
        {
          text: 'Cancel',
          style: 'cancel'
        },
        {
          text: 'Logout',
          onPress: () => {
            // Navigate to login screen
            router.replace('/login');
          },
          style: 'destructive'
        }
      ]
    );
  };

  // Calculate weeks until due date
  const calculateWeeksUntilDueDate = (): number => {
    const currentDate = new Date();
    const dueDate = new Date(USER_PROFILE.dueDate);
    const differenceInTime = dueDate.getTime() - currentDate.getTime();
    const differenceInDays = differenceInTime / (1000 * 3600 * 24);
    return Math.max(0, Math.floor(differenceInDays / 7));
  };

  const weeksUntilDue = calculateWeeksUntilDueDate();

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: 'Profile' }} />
      
      <ScrollView style={styles.scrollView}>
        <View style={styles.profileHeader}>
          <Image 
            source={require('@/assets/images/mother.jpeg')}
            style={styles.profileImage}
            resizeMode="cover"
          />
          <Text style={styles.profileName}>{USER_PROFILE.name}</Text>
          <View style={styles.dueDateBadge}>
            <Text style={styles.dueDateText}>Due in {weeksUntilDue} weeks</Text>
          </View>
          <TouchableOpacity style={styles.editButton}>
            <Text style={styles.editButtonText}>Edit Profile</Text>
          </TouchableOpacity>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Personal Information</Text>
          <View style={styles.infoCard}>
            <View style={styles.infoRow}>
              <View style={styles.infoIconContainer}>
                <IconSymbol size={16} name="phone.fill" color="#5E7CE2" />
              </View>
              <View style={styles.infoContent}>
                <Text style={styles.infoLabel}>Phone Number</Text>
                <Text style={styles.infoValue}>{USER_PROFILE.phoneNumber}</Text>
              </View>
            </View>
            
            {USER_PROFILE.email && (
              <View style={styles.infoRow}>
                <View style={styles.infoIconContainer}>
                  <IconSymbol size={16} name="envelope.fill" color="#5E7CE2" />
                </View>
                <View style={styles.infoContent}>
                  <Text style={styles.infoLabel}>Email</Text>
                  <Text style={styles.infoValue}>{USER_PROFILE.email}</Text>
                </View>
              </View>
            )}
            
            <View style={styles.infoRow}>
              <View style={styles.infoIconContainer}>
                <IconSymbol size={16} name="calendar" color="#5E7CE2" />
              </View>
              <View style={styles.infoContent}>
                <Text style={styles.infoLabel}>Due Date</Text>
                <Text style={styles.infoValue}>{new Date(USER_PROFILE.dueDate).toLocaleDateString()}</Text>
              </View>
            </View>
            
            {USER_PROFILE.bloodType && (
              <View style={styles.infoRow}>
                <View style={styles.infoIconContainer}>
                  <IconSymbol size={16} name="drop.fill" color="#5E7CE2" />
                </View>
                <View style={styles.infoContent}>
                  <Text style={styles.infoLabel}>Blood Type</Text>
                  <Text style={styles.infoValue}>{USER_PROFILE.bloodType}</Text>
                </View>
              </View>
            )}
            
            <View style={styles.infoRow}>
              <View style={styles.infoIconContainer}>
                <IconSymbol size={16} name="location.fill" color="#5E7CE2" />
              </View>
              <View style={styles.infoContent}>
                <Text style={styles.infoLabel}>Region</Text>
                <Text style={styles.infoValue}>{USER_PROFILE.region}</Text>
              </View>
            </View>
          </View>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Healthcare Information</Text>
          <View style={styles.infoCard}>
            <View style={styles.infoRow}>
              <View style={styles.infoIconContainer}>
                <IconSymbol size={16} name="building.2.fill" color="#5E7CE2" />
              </View>
              <View style={styles.infoContent}>
                <Text style={styles.infoLabel}>Health Facility</Text>
                <Text style={styles.infoValue}>{USER_PROFILE.healthFacility}</Text>
              </View>
            </View>
            
            <View style={styles.infoRow}>
              <View style={styles.infoIconContainer}>
                <IconSymbol size={16} name="person.badge.shield.checkmark.fill" color="#5E7CE2" />
              </View>
              <View style={styles.infoContent}>
                <Text style={styles.infoLabel}>Assigned CHW</Text>
                <Text style={styles.infoValue}>{USER_PROFILE.assignedChw}</Text>
              </View>
              <TouchableOpacity style={styles.contactButton}>
                <IconSymbol size={16} name="message.fill" color="#fff" />
              </TouchableOpacity>
            </View>
          </View>
        </View>
        
        {USER_PROFILE.emergencyContact && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Emergency Contact</Text>
            <View style={styles.infoCard}>
              <View style={styles.infoRow}>
                <View style={styles.infoIconContainer}>
                  <IconSymbol size={16} name="person.fill" color="#5E7CE2" />
                </View>
                <View style={styles.infoContent}>
                  <Text style={styles.infoLabel}>Name</Text>
                  <Text style={styles.infoValue}>{USER_PROFILE.emergencyContact.name}</Text>
                </View>
              </View>
              
              <View style={styles.infoRow}>
                <View style={styles.infoIconContainer}>
                  <IconSymbol size={16} name="phone.fill" color="#5E7CE2" />
                </View>
                <View style={styles.infoContent}>
                  <Text style={styles.infoLabel}>Phone Number</Text>
                  <Text style={styles.infoValue}>{USER_PROFILE.emergencyContact.phoneNumber}</Text>
                </View>
                <TouchableOpacity style={styles.contactButton}>
                  <IconSymbol size={16} name="phone.fill" color="#fff" />
                </TouchableOpacity>
              </View>
              
              <View style={styles.infoRow}>
                <View style={styles.infoIconContainer}>
                  <IconSymbol size={16} name="person.2.fill" color="#5E7CE2" />
                </View>
                <View style={styles.infoContent}>
                  <Text style={styles.infoLabel}>Relationship</Text>
                  <Text style={styles.infoValue}>{USER_PROFILE.emergencyContact.relationship}</Text>
                </View>
              </View>
            </View>
          </View>
        )}
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Notification Settings</Text>
          <View style={styles.infoCard}>
            {notificationSettings.map(setting => (
              <View key={setting.id} style={styles.notificationRow}>
                <View style={styles.notificationContent}>
                  <Text style={styles.notificationLabel}>{setting.description}</Text>
                </View>
                <Switch
                  value={setting.enabled}
                  onValueChange={() => toggleNotification(setting.id)}
                  trackColor={{ false: '#d1d1d1', true: '#E91E63' }}
                  thumbColor="#fff"
                />
              </View>
            ))}
          </View>
        </View>
        
        <View style={styles.buttonsSection}>
          <TouchableOpacity 
            style={styles.logoutButton}
            onPress={handleLogout}
          >
            <IconSymbol size={18} name="rectangle.portrait.and.arrow.right" color="#E91E63" />
            <Text style={styles.logoutButtonText}>Logout</Text>
          </TouchableOpacity>
          
          <TouchableOpacity style={styles.sosButton}>
            <IconSymbol size={18} name="exclamationmark.triangle.fill" color="#fff" />
            <Text style={styles.sosButtonText}>Emergency SOS</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  scrollView: {
    flex: 1,
    paddingHorizontal: 16,
  },
  profileHeader: {
    alignItems: 'center',
    marginVertical: 24,
  },
  profileImage: {
    width: 100,
    height: 100,
    borderRadius: 50,
    marginBottom: 16,
    borderWidth: 3,
    borderColor: '#fff', 
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  profileName: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 8,
  },
  dueDateBadge: {
    backgroundColor: '#f8d7da',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 16,
  },
  dueDateText: {
    color: '#721c24',
    fontSize: 14,
    fontWeight: '500',
  },
  editButton: {
    backgroundColor: '#f0f0f0',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
  },
  editButtonText: {
    color: '#555',
    fontSize: 14,
    fontWeight: '500',
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  infoCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  infoIconContainer: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#f0f0f0',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  infoContent: {
    flex: 1,
  },
  infoLabel: {
    fontSize: 14,
    color: '#777',
  },
  infoValue: {
    fontSize: 16,
    color: '#333',
    fontWeight: '500',
  },
  contactButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#5E7CE2',
    justifyContent: 'center',
    alignItems: 'center',
  },
  notificationRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  notificationContent: {
    flex: 1,
  },
  notificationLabel: {
    fontSize: 16,
    color: '#333',
  },
  buttonsSection: {
    marginBottom: 40,
  },
  logoutButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#E91E63',
    borderRadius: 25,
    paddingVertical: 12,
    marginBottom: 16,
  },
  logoutButtonText: {
    color: '#E91E63',
    fontWeight: '600',
    fontSize: 16,
    marginLeft: 8,
  },
  sosButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#EC5766',
    borderRadius: 25,
    paddingVertical: 12,
  },
  sosButtonText: {
    color: '#fff',
    fontWeight: '600',
    fontSize: 16,
    marginLeft: 8,
  },
});
