import React from 'react';
import { View, Text, StyleSheet, SafeAreaView, ScrollView, TouchableOpacity } from 'react-native';
import { Stack } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define interfaces for type safety
interface Appointment {
  id: string;
  date: string;
  time: string;
  doctor: string;
  location: string;
  type: string;
  status: 'confirmed' | 'pending' | 'completed';
}

// Mock data for upcoming appointments
const UPCOMING_APPOINTMENTS: Appointment[] = [
  {
    id: '1',
    date: '2025-05-15',
    time: '10:00 AM',
    doctor: 'Dr. Sarah Williams',
    location: 'Freetown Community Clinic',
    type: 'Prenatal Checkup',
    status: 'confirmed',
  },
  {
    id: '2',
    date: '2025-05-28',
    time: '2:30 PM',
    doctor: 'Dr. Mohamed Conteh',
    location: 'Sierra Leone Maternal Hospital',
    type: 'Ultrasound',
    status: 'pending',
  },
  {
    id: '3',
    date: '2025-06-10',
    time: '9:15 AM',
    doctor: 'Dr. Fatima Bangura',
    location: 'Freetown Community Clinic',
    type: 'Blood Work',
    status: 'confirmed',
  },
];

// Mock data for past appointments
const PAST_APPOINTMENTS: Appointment[] = [
  {
    id: '4',
    date: '2025-04-20',
    time: '11:00 AM',
    doctor: 'Dr. Sarah Williams',
    location: 'Freetown Community Clinic',
    type: 'Initial Consultation',
    status: 'completed',
  },
  {
    id: '5',
    date: '2025-03-15',
    time: '1:45 PM',
    doctor: 'Dr. Mohamed Conteh',
    location: 'Sierra Leone Maternal Hospital',
    type: 'Prenatal Checkup',
    status: 'completed',
  },
];

export default function AppointmentsScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [selectedTab, setSelectedTab] = React.useState<'upcoming' | 'past'>('upcoming');

  const renderAppointmentCard = (appointment: Appointment): React.JSX.Element => {
    const statusColors = {
      confirmed: '#33B47B',
      pending: '#FFAE33',
      completed: '#777777',
    };
    
    const statusBgColors = {
      confirmed: '#DFF7E9',
      pending: '#FFF6DD',
      completed: '#F0F0F0',
    };

    const dateObj = new Date(appointment.date);
    const formattedDate = dateObj.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    });

    return (
      <TouchableOpacity
        key={appointment.id}
        style={styles.appointmentCard}
        onPress={() => {
          // Navigate to appointment details in the future
        }}
      >
        <View style={styles.appointmentHeader}>
          <View style={styles.dateContainer}>
            <Text style={styles.dateText}>{formattedDate}</Text>
            <Text style={styles.timeText}>{appointment.time}</Text>
          </View>
          <View 
            style={[
              styles.statusBadge, 
              { backgroundColor: statusBgColors[appointment.status] }
            ]}
          >
            <Text 
              style={[
                styles.statusText, 
                { color: statusColors[appointment.status] }
              ]}
            >
              {appointment.status.charAt(0).toUpperCase() + appointment.status.slice(1)}
            </Text>
          </View>
        </View>
        
        <View style={styles.appointmentDetails}>
          <View style={styles.detailRow}>
            <IconSymbol size={16} name="stethoscope" color="#555" />
            <Text style={styles.detailText}>{appointment.type}</Text>
          </View>
          <View style={styles.detailRow}>
            <IconSymbol size={16} name="person.fill" color="#555" />
            <Text style={styles.detailText}>{appointment.doctor}</Text>
          </View>
          <View style={styles.detailRow}>
            <IconSymbol size={16} name="mappin.and.ellipse" color="#555" />
            <Text style={styles.detailText}>{appointment.location}</Text>
          </View>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: 'Appointments' }} />
      
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Appointments</Text>
        <TouchableOpacity style={styles.addButton}>
          <IconSymbol size={24} name="plus" color="#fff" />
        </TouchableOpacity>
      </View>
      
      <View style={styles.tabContainer}>
        <TouchableOpacity 
          style={[
            styles.tabButton, 
            selectedTab === 'upcoming' && styles.selectedTab
          ]}
          onPress={() => setSelectedTab('upcoming')}
        >
          <Text 
            style={[
              styles.tabText, 
              selectedTab === 'upcoming' && styles.selectedTabText
            ]}
          >
            Upcoming
          </Text>
        </TouchableOpacity>
        <TouchableOpacity 
          style={[
            styles.tabButton, 
            selectedTab === 'past' && styles.selectedTab
          ]}
          onPress={() => setSelectedTab('past')}
        >
          <Text 
            style={[
              styles.tabText, 
              selectedTab === 'past' && styles.selectedTabText
            ]}
          >
            Past
          </Text>
        </TouchableOpacity>
      </View>
      
      <ScrollView style={styles.scrollView}>
        {selectedTab === 'upcoming' ? (
          UPCOMING_APPOINTMENTS.map(renderAppointmentCard)
        ) : (
          PAST_APPOINTMENTS.map(renderAppointmentCard)
        )}
        
        {selectedTab === 'upcoming' && UPCOMING_APPOINTMENTS.length === 0 && (
          <View style={styles.emptyState}>
            <IconSymbol size={64} name="calendar.badge.exclamationmark" color="#ccc" />
            <Text style={styles.emptyStateText}>No upcoming appointments</Text>
            <TouchableOpacity style={styles.scheduleButton}>
              <Text style={styles.scheduleButtonText}>Schedule an Appointment</Text>
            </TouchableOpacity>
          </View>
        )}
        
        {selectedTab === 'past' && PAST_APPOINTMENTS.length === 0 && (
          <View style={styles.emptyState}>
            <IconSymbol size={64} name="calendar.badge.clock" color="#ccc" />
            <Text style={styles.emptyStateText}>No past appointments</Text>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 16,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#333',
  },
  addButton: {
    backgroundColor: '#E91E63',
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  tabContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    marginBottom: 16,
  },
  tabButton: {
    paddingVertical: 10,
    paddingHorizontal: 16,
    marginRight: 10,
    borderRadius: 20,
  },
  selectedTab: {
    backgroundColor: '#E91E63',
  },
  tabText: {
    fontSize: 16,
    color: '#555',
  },
  selectedTabText: {
    color: '#fff',
    fontWeight: '500',
  },
  scrollView: {
    paddingHorizontal: 16,
  },
  appointmentCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  appointmentHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  dateContainer: {
    flexDirection: 'column',
  },
  dateText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  timeText: {
    fontSize: 14,
    color: '#777',
    marginTop: 4,
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '500',
  },
  appointmentDetails: {
    marginTop: 8,
  },
  detailRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  detailText: {
    fontSize: 14,
    color: '#555',
    marginLeft: 8,
  },
  emptyState: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: 40,
  },
  emptyStateText: {
    fontSize: 16,
    color: '#777',
    marginTop: 16,
  },
  scheduleButton: {
    backgroundColor: '#E91E63',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 24,
    marginTop: 16,
  },
  scheduleButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '500',
  },
});
